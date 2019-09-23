FROM golang:1.13-alpine as builder

# Setup
RUN mkdir -p /go/src/github.com/thomseddon/udp-replicator
WORKDIR /go/src/github.com/thomseddon/udp-replicator

# Add libraries
RUN apk add --no-cache git

# Copy & build
ADD . /go/src/github.com/thomseddon/udp-replicator/
RUN CGO_ENABLED=0 GOOS=linux GO111MODULE=on go build -a -installsuffix nocgo -o /udp-replicator github.com/thomseddon/udp-replicator

# Copy into scratch container
FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /udp-replicator ./
ENTRYPOINT ["./udp-replicator"]
