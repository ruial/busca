package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/ruial/busca/internal/api"
	"github.com/ruial/busca/internal/repository"
)

func exportIndexRepo(indexRepo repository.IndexRepo) {
	log.Println("Exporting index")
	start := time.Now()
	if err := indexRepo.SnapshotExport(); err != nil {
		log.Println(err)
	}
	log.Println("Elapsed exporting:", time.Now().Sub(start))
}

func importIndexRepo(indexRepo repository.IndexRepo) {
	log.Println("Importing index")
	start := time.Now()
	if err := indexRepo.SnapshotImport(); err != nil {
		log.Println(err)
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

	var indexRepo repository.IndexRepo
	indexRepo = &repository.LocalIndexRepo{
		SnapshotDir:      *dataDir,
		SnapshotInterval: snapshotIntervalDuration,
	}
	importIndexRepo(indexRepo)

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
	if indexRepo.IsSnapshotEnabled() {
		wg.Add(1)
		ticker := time.NewTicker(snapshotIntervalDuration)
		defer ticker.Stop()
		go func() {
			for {
				select {
				case <-ticker.C:
					exportIndexRepo(indexRepo)
				case <-done:
					log.Println("Stopping snapshots")
					exportIndexRepo(indexRepo)
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
