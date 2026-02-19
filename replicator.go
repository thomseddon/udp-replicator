package main

import (
	"fmt"
	"net"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	debug      = kingpin.Flag("debug", "Enable debug mode").Envar("DEBUG").Bool()
	listenIP   = kingpin.Flag("listen-ip", "IP to listen in").Default("0.0.0.0").Envar("LISTEN_IP").IP()
	listenPort = kingpin.Flag("listen-port", "Port to listen on").Default("9000").Envar("LISTEN_PORT").Int()
	bodySize   = kingpin.Flag("body-size", "Size of body to read").Default("4096").Envar("BODY_SIZE").Int()
	ttl        = kingpin.Flag("ttl", "TTL for forwarded packets (0 = OS default)").Default("0").Envar("TTL").Int()

	forwards = kingpin.Flag("forward", "ip:port to forward traffic to (port defaults to listen-port)").PlaceHolder("ip:port").Envar("FORWARD").Strings()

	pretty = kingpin.Flag("pretty", "").Default("true").Envar("PRETTY").Hidden().Bool()

	targets []*net.UDPConn
)

func main() {
	// CLI
	kingpin.Parse()

	// Log setup
	if *debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
	if !*pretty {
		log.SetFormatter(&log.TextFormatter{
			DisableColors: true,
			FullTimestamp: true,
		})
	}

	if len(*forwards) <= 0 {
		log.Fatal("Must specify at least one forward target")
	}

	// Clients
	for _, forward := range *forwards {
		// Check for port
		if strings.Index(forward, ":") < 0 {
			forward = fmt.Sprintf("%s:%d", forward, *listenPort)
		}

		// Resolve
		addr, err := net.ResolveUDPAddr("udp", forward)
		if err != nil {
			log.Fatalf("Could not ResolveUDPAddr: %s (%s)", forward, err)
		}

		// Setup conn
		conn, err := net.DialUDP("udp", nil, addr)
		if err != nil {
			log.Fatalf("Could not DialUDP: %+v (%s)", addr, err)
		}
		defer conn.Close()

		// Set TTL if specified
		if *ttl > 0 {
			rawConn, err := conn.SyscallConn()
			if err != nil {
				log.Fatalf("Could not get raw conn for %s: %s", forward, err)
			}
			var sysErr error
			rawConn.Control(func(fd uintptr) {
				if err := syscall.SetsockoptInt(int(fd), syscall.IPPROTO_IP, syscall.IP_TTL, *ttl); err != nil {
					sysErr = err
					return
				}
				if err := syscall.SetsockoptInt(int(fd), syscall.IPPROTO_IP, syscall.IP_MULTICAST_TTL, *ttl); err != nil {
					sysErr = err
					return
				}
			})
			if sysErr != nil {
				log.Fatalf("Could not set TTL on %s: %s", forward, sysErr)
			}
		}

		targets = append(targets, conn)
	}

	// Server
	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		Port: *listenPort,
		IP:   *listenIP,
	})
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	// Startup status
	log.WithFields(log.Fields{
		"ip":   *listenIP,
		"port": *listenPort,
		"ttl":  *ttl,
	}).Infof("Server started")
	for i, target := range targets {
		log.WithFields(log.Fields{
			"num":   i + 1,
			"total": len(targets),
			"addr":  target.RemoteAddr(),
		}).Info("Forwarding target configured")
	}

	for {
		// Read
		b := make([]byte, *bodySize)
		n, addr, err := conn.ReadFromUDP(b)
		if err != nil {
			log.Error(err)
			continue
		}

		// Log receive
		ctxLog := log.WithFields(log.Fields{
			"source": addr.String(),
			"body":   string(b[:n]),
		})
		ctxLog.Debugf("Recieved packet")

		// Proxy
		for _, target := range targets {
			_, err := target.Write(b[:n])

			// Log proxy
			ctxLog := ctxLog.WithFields(log.Fields{
				"target": target.RemoteAddr(),
			})

			if err != nil {
				ctxLog.Warn("Could not forward packet", err)
			} else {
				ctxLog.Debug("Wrote to target")
			}
		}
	}
}
