### STAGE 1: Build ###

FROM golang:buster as builder

RUN mkdir -p /app/src/github.com/DRuggeri/netgear_exporter
ENV GOPATH /app
WORKDIR /app
COPY . /app/src/github.com/DRuggeri/netgear_exporter
RUN go install github.com/DRuggeri/netgear_exporter

### STAGE 2: Setup ###

FROM alpine
RUN apk add --no-cache \
  libc6-compat
COPY --from=builder /app/bin/netgear_exporter /netgear_exporter
RUN chmod +x /netgear_exporter
ENTRYPOINT ["/netgear_exporter"]
