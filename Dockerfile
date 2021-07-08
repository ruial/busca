FROM golang:1.16 AS builder
WORKDIR /build
COPY . .
RUN go get -d -v ./...
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o app ./cmd/busca

FROM scratch
COPY --from=builder /build/app .
ENTRYPOINT ["./app"]
CMD ["--data-dir=data"]
