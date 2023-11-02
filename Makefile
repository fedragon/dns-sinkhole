.PHONY: deploy test

GOOS ?= linux
GOARCH ?= arm
GOARM ?= 6

HOSTS_URL ?= https://raw.githubusercontent.com/StevenBlack/hosts/master/alternates/fakenews-gambling-porn/hosts

RPI_HOST ?= raspberrypi
RPI_USER ?= pi

build:
	GOOS=${GOOS} GOARCH=${GOARCH} GOARM=${GOARM} go build -o deploy/hole cmd/main.go

fetch:
	HOSTS_URL="${HOSTS_URL}" curl -sSL "${HOSTS_URL}" -o deploy/hosts

test:
	go test -v -race -count=1 ./...

generate-service:
	@mkdir -p deploy
	@RPI_USER=${RPI_USER} envsubst < templates/sinkhole.service > deploy/sinkhole.service

deploy:
	RPI_HOST=${RPI_HOST} RPI_USER=${RPI_USER} scp deploy/* "${RPI_USER}@${RPI_HOST}":
