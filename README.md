# sinkhole

**Toy** project to explore DNS Sinkholes. Acts as a [DNS sinkhole](https://en.wikipedia.org/wiki/DNS_sinkhole), receiving DNS queries and returning non-routable addresses for blacklisted domains; legitimate DNS queries are forwarded to a (configurable) upstream DNS resolver.

It can (optionally) expose an HTTP endpoint to provide metrics to a Prometheus server (see configuration options).

## Motivation

One day I started reading the [Running pi-hole on a Raspberry Pi](https://www.raspberrypi.com/tutorials/running-pi-hole-on-a-raspberry-pi/) tutorial and I thought that writing my own DNS sinkhole could be an interesting pet project, so here we are.

**Note:** While I've developed it with a Raspberry Pi in mind (simply because I had an old one lying in a drawer at home), it should in principle work on any other Linux machine that uses `systemd` (remember to change the target architecture accordingly, when building the binary).

## Current limitations

It can currently only resolve queries received over UDP for: 

- A-type or AAAA-type (IPv4 or IPv6)
- IN-class 

Any other query will be forwarded to the upstream DNS resolver.

## Usage

Choose your preferred version of Steven Black's Hosts [here](https://github.com/StevenBlack/hosts#list-of-all-hosts-file-variants), then run

### 1. Download hosts file

```shell
HOSTS_URL="<link to your chosen version of Steven Black Hosts file>" make fetch
```

### 2. Build

```shell
# build executable binary
GOOS=linux GOARCH=arm GOARM=6 make build

# generate systemd service pointing to an executable in the user's home directory
RPI_USER=<user> METRICS_ENABLED=<true|false> AUDIT_LOG_ENABLED=<true|false> make generate-service
```

### 3. Deploy to your target server (e.g. a Raspberry Pi)

```shell
RPI_HOST=<host> RPI_USER=<user> make deploy
```

### 4. Install

**Note:** This script runs with privileged permissions, so make sure to inspect it beforehand!

```shell
sudo ./install

sudo systemctl start sinkhole.service

# tail service logs to check if it's working as intended, then quit if everything is okay
sudo journalctl -f -u sinkhole.service
```

## Uninstall

**Note:** This script runs with privileged permissions, so make sure to inspect it beforehand!

```shell
sudo ./uninstall
```

## Local development

### Run

```shell
# note: this command uses the following defaults:
# LOCAL_SERVER_ADDR="0.0.0.0:53"    # address of the UDP server used to receive DNS queries
# UPSTREAM_SERVER_ADDR="1.1.1.1:53" # DNS recursive resolver for legitimate queries (default: Cloudflare's)
# HOSTS_PATH="./hosts"              # path to the hosts file containing blacklisted domains
# METRICS_ENABLED="true"            # expose endpoint for Prometheus metrics?
# HTTP_SERVER_ADDR="0.0.0.0:8000"   # address of the HTTP server (only started if METRICS_ENABLED=true)
# overwrite any of them if/as needed using environment variables

deploy/hole
```
