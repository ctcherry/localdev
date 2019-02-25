FROM golang:1.11

WORKDIR /app/src

COPY *.go ./

RUN go build -o tlself

FROM debian:stretch

COPY --from=0 /app/src/tlself /app/

ENV CERTDIR=/certs

ENTRYPOINT ["/app/tlself"]
