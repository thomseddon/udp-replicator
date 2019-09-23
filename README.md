
# UDP Replicator [![Go Report Card](https://goreportcard.com/badge/github.com/thomseddon/udp-replicator)](https://goreportcard.com/report/github.com/thomseddon/udp-replicator)

A tiny UDP proxy that can replicate traffic to one or more endpoints.

## Why?

We needed a way to take a single netflow stream and send it to multiple endpoints. As this is a generic UDP replicator, it can be used for any traffic such as netflow, syslog etc.

You could use iptables, [like Zapier](https://zapier.com/engineering/iptables-replication/), but we wanted something we could easily deploy into our kubernetes cluster. There are also a number of existing UDP proxies, but non of the popular ones support replication. There are also a small number of replicators, but these all seem to be mostly unmaintained/untested toy projects with little documentation.

## Usage

### Direct

```
./replicator --listen-port=9500 --forward=192.0.2.1 --forward=198.51.100.5:9100
time="2019-09-23T14:29:50+01:00" level=info msg="Server started" ip=0.0.0.0 port=9500
time="2019-09-23T14:29:50+01:00" level=info msg="Forwarding target configured" addr="192.0.2.1:9500" num=1 total=2
time="2019-09-23T14:29:50+01:00" level=info msg="Forwarding target configured" addr="198.51.100.5:9100" num=2 total=2
```

The above command will:

- Start a UDP server listening on `0.0.0.0` (the default) port `9500`
- Add a forward target of `192.0.2.1:9500` (uses `listen-port` for destination port as not specified in configuration)
- Add another forward target of `198.51.100.5:9100`

The server will start listening on `0.0.0.0:9500`, any packet it receives will be replicated and sent to both `192.0.2.1:9500` and `198.51.100.5:9100`


### Docker

```
docker run -e FORWARDS=$'192.0.2.1\n198.51.100.5:9100' thomseddon/udp-replicator:1
```

### Docker Compose

```
version: '3'

services:
  udp-replicator:
    run: thomseddon/udp-replicator:1
    environment:
      DEBUG: "true"
      FORWARD: "192.0.2.1\n198.51.100.5:9100"
```


## Configuration

```
usage: replicator [<flags>]

Flags:
  --help                 Show context-sensitive help (also try --help-long and --help-man).
  --debug                Enable debug mode
  --listen-ip=0.0.0.0    IP to listen in
  --listen-port=9000     Port to listen on
  --body-size=4096       Size of body to read
  --forward=ip:port ...  ip:port to forward traffic to (port defaults to listen-port)
```

All configuration params can be passed as arguments as shown above, or as environment varaibles (as shown in usage above) - where environment variables are uppercased snake case e.g. `LISTEN_IP`

## Copyright

2019 Thom Seddon

## License

[MIT](https://github.com/thomseddon/udp-replicator/blob/master/LICENSE.md)
