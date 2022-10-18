FROM docker.io/library/alpine:3.16.2 as certs
RUN apk add --update --no-cache ca-certificates

FROM docker.io/library/busybox:1.35.0
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/bin/git-ops-update"]
COPY git-ops-update /bin/git-ops-update
WORKDIR /workdir
