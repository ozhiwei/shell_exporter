FROM golang:1.14.1-alpine AS build-env

RUN apk add --update git gcc libc-dev
RUN go get -u github.com/prometheus/promu

RUN mkdir shell_exporter
COPY .promu.yml shell_exporter.go go.mod go.sum /go/shell_exporter/

WORKDIR /go/shell_exporter
RUN promu build

FROM alpine:3.11
RUN apk add --no-cache bash
COPY --from=build-env /go/shell_exporter/shell_exporter /bin/shell_exporter
COPY config.yml /etc/shell_exporter/config.yml

EXPOSE      9191
ENTRYPOINT  [ "/bin/shell_exporter" ]
CMD ["-config.file=/etc/shell_exporter/config.yml"]
