# Use the official Golang image as the base image
FROM golang:1.22

COPY ./go-app /app/go
WORKDIR /app/go

# Download all dependencies
RUN go mod download

RUN go install github.com/air-verse/air@latest
