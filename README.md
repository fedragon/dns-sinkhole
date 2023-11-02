# sinkhole

Acts as a [DNS sinkhole](https://en.wikipedia.org/wiki/DNS_sinkhole), receiving DNS queries and returning non-routable addresses for blacklisted domains; legitimate DNS queries are forwarded to a (configurable) upstream DNS resolver.

It can (optionally) expose an HTTP endpoint to provide metrics to a Prometheus server (see configuration options).

## Motivation

One day I started reading the [Running pi-hole on a Raspberry Pi](https://www.raspberrypi.com/tutorials/running-pi-hole-on-a-raspberry-pi/) tutorial and I thought that writing my own DNS sinkhole could be an interesting pet project, so here we are.

## Current limitations

This service can currently only resolve A-type, IN-class queries received over UDP: any other query will be forwarded to the upstream DNS resolver.

## Usage

Choose your preferred version of Steven Black's Hosts [here](https://github.com/StevenBlack/hosts#list-of-all-hosts-file-variants), then run

### 1. Download hosts file

```shell
HOSTS_URL="<link to your chosen version of Steven Black Hosts file>" make fetch
```

### 2. Build

```shell
# note: this command uses the following defaults:
# GOOS=linux
# GOARCH=arm
# GOARM=6
# overwrite any of them if/as needed using environment variables

make build
```

### 3. Deploy to your target server (e.g. a Raspberry Pi)

```shell
RPI_USER=pi make generate-service               # generates systemd service pointing to an executable in the user's home directory
RPI_HOST=raspberrypi RPI_USER=pi make deploy    # copies the required files to the RPi via scp```
```

### 3. Run

```shell
# note: this command uses the following defaults:
# LOCAL_SERVER_ADDR="0.0.0.0:1153"    # address of the UDP server used to receive DNS queries
# UPSTREAM_SERVER_ADDR="1.1.1.1:53"   # DNS recursive resolver for legitimate queries (default: Cloudflare's)
# HOSTS_PATH="./hosts"              # path to the hosts file containing blacklisted domains
# METRICS_ENABLED="true"            # expose endpoint for Prometheus metrics?
# HTTP_SERVER_ADDR="0.0.0.0:8000"   # address of the HTTP server (only started if METRICS_ENABLED=true)
# overwrite any of them if/as needed using environment variables

bin/sinkhole
```
