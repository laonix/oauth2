FROM golang:alpine AS builder

WORKDIR /build

COPY . .

RUN go build -o oauth2 cmd/main.go

FROM alpine

WORKDIR /build

COPY --from=builder /build/oauth2 /build/oauth2
COPY --from=builder /build/config.yaml /build/config.yaml

EXPOSE 3000

CMD ["./oauth2"]
