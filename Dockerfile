FROM golang:1.14-alpine AS builder

WORKDIR /build
COPY ./ /build/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o git-ops-update ./src

FROM alpine
COPY --from=builder /build/git-ops-update /bin/git-ops-update
WORKDIR /workdir
ENTRYPOINT ["/bin/git-ops-update"]
