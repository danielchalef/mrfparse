# syntax=docker/dockerfile:1
FROM golang:1.19 AS build

WORKDIR /app

COPY go.mod .
COPY go.sum .
COPY main.go .
RUN go mod download

COPY cmd cmd
COPY pkg pkg

RUN GOOS=linux GOARCH=amd64 GOAMD64=v3 go build -ldflags="-w -s" -o /mrfparse 

FROM debian:bullseye-slim as runtime

RUN mkdir /app
WORKDIR /app

COPY --from=build /mrfparse .
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

COPY config.yaml .
COPY data/tic_500_shoppable_services.csv services.csv

ENTRYPOINT [ "/app/mrfparse" ]