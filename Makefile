# https://danishpraka.sh/2019/12/07/using-makefiles-for-go.html

build:
	go build -o busca.out cmd/busca/main.go

run:
	go run cmd/busca/main.go -snapshot-interval=60s -data-dir=data

test:
	go test ./...

test-unit:
	go test ./... -short

test-race:
	go test -v -race ./...

test-bench:
	go test ./pkg/index -bench . -benchtime=5000x -benchmem

test-all: test-race test-bench

# https://golang.org/pkg/net/http/pprof/
# https://blog.golang.org/pprof
test-profile:
	go test ./pkg/index -bench . -benchtime=5000x -benchmem -cpuprofile profile.out
	go tool pprof -web profile.out

cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out

bootstrap:
	./scripts/loader.sh
