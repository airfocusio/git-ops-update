FROM golang:1.17-alpine AS builder

WORKDIR /build
COPY ./ /build/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o git-ops-update .

FROM alpine:latest as certs
RUN apk add --update --no-cache ca-certificates

FROM scratch
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/git-ops-update /bin/git-ops-update
WORKDIR /workdir
ENTRYPOINT ["/bin/git-ops-update"]
