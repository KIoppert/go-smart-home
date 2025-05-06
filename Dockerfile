FROM ubuntu:latest
FROM golang:latest

COPY . /app
WORKDIR /app

ENTRYPOINT go test ./... -v -race