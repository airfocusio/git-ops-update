FROM alpine:3.16.0 as certs
RUN apk add --update --no-cache ca-certificates

FROM busybox:1.34.1
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/bin/git-ops-update"]
COPY git-ops-update /bin/git-ops-update
WORKDIR /workdir
