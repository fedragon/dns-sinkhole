build: export GOOS=linux
build: export GOARCH=arm
build: export GOARM=6
build:
	mkdir -p bin
	go build -o bin/sinkhole cmd/main.go
