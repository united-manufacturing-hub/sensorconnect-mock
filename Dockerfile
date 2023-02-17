FROM docker.io/golang:1.19-alpine3.16 as builder

WORKDIR /build

RUN apk add build-base

COPY go.mod go.sum ./

RUN go mod download && go mod verify

COPY main.go utils.go ./

RUN GOOS=linux go build -tags musl -ldflags "--extldflags '-static'" -o app ./...

FROM alpine:latest as certs

RUN apk --update add ca-certificates

RUN apk add build-base
#FROM scratch
#
#COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
#
COPY --from=builder /build/app /app

CMD ["./app"]