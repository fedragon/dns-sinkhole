.PHONY: deploy test

GOOS ?= linux
GOARCH ?= arm
GOARM ?= 6

HOSTS_URL ?= https://raw.githubusercontent.com/StevenBlack/hosts/master/alternates/fakenews-gambling-porn/hosts

RPI_HOST ?= raspberrypi
RPI_USER ?= pi

METRICS_ENABLED ?= false
AUDIT_LOG_ENABLED ?= false

VERSION ?= $(shell date +%Y%m%dT%H%M%S)

all: build fetch generate-service deploy

pre:
	@mkdir -p deploy

build: pre
	@echo "Building version ${VERSION}"
	@GOOS=${GOOS} GOARCH=${GOARCH} GOARM=${GOARM} go build -ldflags="-X 'main.Version=${VERSION}'" -o deploy/hole cmd/main.go

fetch: pre
	HOSTS_URL="${HOSTS_URL}" curl -sSL "${HOSTS_URL}" -o deploy/hosts

test:
	go test -v -race -count=1 ./...

generate-service: pre
	@RPI_USER=${RPI_USER} METRICS_ENABLED=${METRICS_ENABLED} AUDIT_LOG_ENABLED=${AUDIT_LOG_ENABLED} envsubst < templates/sinkhole.service > deploy/sinkhole.service

deploy:
	RPI_HOST=${RPI_HOST} RPI_USER=${RPI_USER} scp deploy/* "${RPI_USER}@${RPI_HOST}":
