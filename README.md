# sinkhole

Acts as a DNS sinkhole, receiving DNS queries and returning non-routable addresses for blacklisted domains. It resolves legitimate DNS queries using a (configurable) fallback DNS server.

Like [pi-hole](https://github.com/pi-hole/pi-hole), but without most of its features 😆.

## Motivation

One day I started reading the [Running pi-hole on a Raspberry Pi](https://www.raspberrypi.com/tutorials/running-pi-hole-on-a-raspberry-pi/) tutorial... and writing my own DNS sinkhole sounded like an interesting pet project, so here we are.

## Usage

Choose your preferred version of Steven Black's Hosts [here](https://github.com/StevenBlack/hosts#list-of-all-hosts-file-variants), then run

```shell
HOSTS_URL=<link to your chosen version of Steven Black Hosts file> make fetch

make run
```
