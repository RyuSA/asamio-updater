FROM golang:1.20.2-alpine as builder
WORKDIR /app
COPY go.mod go.sum /app
RUN go mod download
COPY internal/ /app/internal/
COPY main.go /app
RUN go build -o bin/main main.go

FROM ubuntu:22.04 as env
RUN apt-get update \
 && apt-get install -y --force-yes --no-install-recommends apt-transport-https curl ca-certificates \
 && apt-get clean \
 && apt-get autoremove \
 && rm -rf /var/lib/apt/lists/* 

FROM env
WORKDIR /app
COPY --from=builder /app/bin/ /app/
ENTRYPOINT ["/app/main"]
CMD ["-h"]
