FROM golang:alpine AS builder

WORKDIR /build

COPY . .

RUN go build -o oauth cmd/main.go

FROM alpine

WORKDIR /build

COPY --from=builder /build/oauth /build/oauth
COPY --from=builder /build/config.yaml /build/config.yaml

EXPOSE 3000

CMD ["./oauth"]
