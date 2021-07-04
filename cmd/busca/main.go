package main

import (
	"context"
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/ruial/busca/internal/api"
	"github.com/ruial/busca/internal/repository"
	"github.com/ruial/busca/pkg/index"
)

const indexExtension = ".out"

func exportIndexRepo(dir string, indexRepo repository.IndexRepo) {
	log.Println("Exporting index")
	start := time.Now()
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.Mkdir(dir, 0700); err != nil {
			log.Println("Error creating export directory:", err)

		}
	}
	for _, idx := range indexRepo.GetIndexes() {
		out := path.Join(dir, idx.ID+indexExtension)
		// cloning is faster than serializing, so lock time is reduced for readers
		if err := index.Export(idx.Index.Clone(), out); err != nil {
			log.Printf("Error exporting index %s: %s", idx.ID, err.Error())
		}
	}
	log.Println("Elapsed exporting:", time.Now().Sub(start))
}

func importIndexRepo(dir string, indexRepo repository.IndexRepo) {
	log.Println("Importing index")
	start := time.Now()
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Println("Error reading directory:", err)
	}
	for _, file := range files {
		id := strings.TrimSuffix(file.Name(), indexExtension)
		idx, err := index.Import(path.Join(dir, file.Name()))
		if err != nil {
			log.Printf("Error importing index %s: %s", id, err.Error())
		}
		indexRepo.CreateIndex(repository.IdentifiableIndex{ID: id, Index: idx})
	}
	log.Println("Elapsed importing:", time.Now().Sub(start))
}

func main() {
	port := flag.String("port", "8080", "http port")
	address := flag.String("addr", "", "address")
	dataDir := flag.String("data-dir", "", "data directory")
	snapshotInterval := flag.String("snapshot-interval", "", "snapshot interval, e.g.: 60s")
	flag.Parse()

	var snapshotIntervalDuration time.Duration
	if *snapshotInterval != "" {
		if *dataDir == "" {
			log.Fatalln("The data-dir flag must be specified when setting snapshot interval")
		}
		duration, err := time.ParseDuration(*snapshotInterval)
		if err != nil {
			log.Fatalln("Invalid snapshot interval:", *snapshotInterval)
		}
		if duration < 10*time.Second {
			log.Fatalln("Snapshot interval must be higher or equal to 10 seconds")
		}
		snapshotIntervalDuration = duration
	}

	indexRepo := repository.NewInMemoryIndexRepo()
	if *dataDir != "" {
		importIndexRepo(*dataDir, indexRepo)
	}

	addr := *address + ":" + *port
	router := api.SetupRouter(indexRepo)
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	go func() {
		log.Println("Starting server on addr", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalln("Server error:", err)
		} else {
			log.Println("Server graceful shutdown")
		}
	}()

	var wg sync.WaitGroup
	done := make(chan struct{})
	if *snapshotInterval != "" {
		wg.Add(1)
		ticker := time.NewTicker(snapshotIntervalDuration)
		defer ticker.Stop()
		go func() {
			for {
				select {
				case <-ticker.C:
					exportIndexRepo(*dataDir, indexRepo)
				case <-done:
					log.Println("Stopping snapshots")
					exportIndexRepo(*dataDir, indexRepo)
					wg.Done()
					return
				}
			}
		}()
	}

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit

	// graceful shutdown with 3 seconds timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalln("Server forced to shutdown:", err)
	}

	// wait for final snapshot, they are done from single thread so no concurrency issues
	close(done)
	wg.Wait()
	log.Println("bye")
}
