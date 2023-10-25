build: export GOOS=linux
build: export GOARCH=arm
build: export GOARM=6
build:
	mkdir -p bin
	go build -o bin/sinkhole cmd/main.go

fetch: export HOSTS_URL=https://raw.githubusercontent.com/StevenBlack/hosts/master/alternates/fakenews-gambling-porn-social/hosts
fetch:
	curl -sSL "${HOSTS_URL}" -o hosts
