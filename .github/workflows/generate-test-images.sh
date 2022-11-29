#!/bin/bash
set -euo pipefail

export REPOSITORY=ghcr.io/airfocusio/git-ops-update-test

# docker v2 manifest
echo -e "FROM scratch\nLABEL org.opencontainers.image.version=0.0.0\n" | podman build -f - --format=docker --timestamp=0 --no-cache --tag ${REPOSITORY}:docker-v2-manifest-v0.0.0
echo -e "FROM scratch\nLABEL org.opencontainers.image.version=0.0.1\nLABEL org.opencontainers.image.source=https://github.com/airfocusio/git-ops-update\nLABEL org.opencontainers.image.revision=99a7aaa2eff5aff7c77d9caf0901454fedf6bf00\n" | podman build -f - --format=docker --timestamp=0 --no-cache --tag ${REPOSITORY}:docker-v2-manifest-v0.0.1
echo -e "FROM scratch\nLABEL org.opencontainers.image.version=0.0.2\nLABEL org.opencontainers.image.source=https://github.com/airfocusio/git-ops-update\nLABEL org.opencontainers.image.revision=7f4304f54fd1f89aecc15ec3e70975e5bacfeb68\n" | podman build -f - --format=docker --timestamp=0 --no-cache --tag ${REPOSITORY}:docker-v2-manifest-v0.0.2
podman push ${REPOSITORY}:docker-v2-manifest-v0.0.0
podman push ${REPOSITORY}:docker-v2-manifest-v0.0.1
podman push ${REPOSITORY}:docker-v2-manifest-v0.0.2
podman image rm ${REPOSITORY}:docker-v2-manifest-v0.0.0
podman image rm ${REPOSITORY}:docker-v2-manifest-v0.0.1
podman image rm ${REPOSITORY}:docker-v2-manifest-v0.0.2

# oci v1 manifest
echo -e "FROM scratch\nLABEL org.opencontainers.image.version=0.0.0\n" | podman build -f - --format=oci --timestamp=0 --no-cache --tag ${REPOSITORY}:oci-v1-manifest-v0.0.0
echo -e "FROM scratch\nLABEL org.opencontainers.image.version=0.0.1\nLABEL org.opencontainers.image.source=https://github.com/airfocusio/git-ops-update\nLABEL org.opencontainers.image.revision=99a7aaa2eff5aff7c77d9caf0901454fedf6bf00\n" | podman build -f - --format=oci --timestamp=0 --no-cache --tag ${REPOSITORY}:oci-v1-manifest-v0.0.1
echo -e "FROM scratch\nLABEL org.opencontainers.image.version=0.0.2\nLABEL org.opencontainers.image.source=https://github.com/airfocusio/git-ops-update\nLABEL org.opencontainers.image.revision=7f4304f54fd1f89aecc15ec3e70975e5bacfeb68\n" | podman build -f - --format=oci --timestamp=0 --no-cache --tag ${REPOSITORY}:oci-v1-manifest-v0.0.2
podman push ${REPOSITORY}:oci-v1-manifest-v0.0.0
podman push ${REPOSITORY}:oci-v1-manifest-v0.0.1
podman push ${REPOSITORY}:oci-v1-manifest-v0.0.2
podman image rm ${REPOSITORY}:oci-v1-manifest-v0.0.0
podman image rm ${REPOSITORY}:oci-v1-manifest-v0.0.1
podman image rm ${REPOSITORY}:oci-v1-manifest-v0.0.2

# docker v2 manifest list
echo -e "FROM scratch\nLABEL org.opencontainers.image.version=0.0.0\n" | podman build -f - --no-cache --timestamp=0 --platform=linux/amd64 --manifest ${REPOSITORY}:docker-v2-manifest-list-v0.0.0
echo -e "FROM scratch\nLABEL org.opencontainers.image.version=0.0.1\nLABEL org.opencontainers.image.source=https://github.com/airfocusio/git-ops-update\nLABEL org.opencontainers.image.revision=99a7aaa2eff5aff7c77d9caf0901454fedf6bf00\n" | podman build -f - --no-cache --timestamp=0 --platform=linux/amd64 --manifest ${REPOSITORY}:docker-v2-manifest-list-v0.0.1
echo -e "FROM scratch\nLABEL org.opencontainers.image.version=0.0.2\nLABEL org.opencontainers.image.source=https://github.com/airfocusio/git-ops-update\nLABEL org.opencontainers.image.revision=7f4304f54fd1f89aecc15ec3e70975e5bacfeb68\n" | podman build -f - --no-cache --timestamp=0 --platform=linux/amd64 --manifest ${REPOSITORY}:docker-v2-manifest-list-v0.0.2
podman manifest push --format=v2s2 --all ${REPOSITORY}:docker-v2-manifest-list-v0.0.0 docker://${REPOSITORY}:docker-v2-manifest-list-v0.0.0
podman manifest push --format=v2s2 --all ${REPOSITORY}:docker-v2-manifest-list-v0.0.1 docker://${REPOSITORY}:docker-v2-manifest-list-v0.0.1
podman manifest push --format=v2s2 --all ${REPOSITORY}:docker-v2-manifest-list-v0.0.2 docker://${REPOSITORY}:docker-v2-manifest-list-v0.0.2
podman manifest rm ${REPOSITORY}:docker-v2-manifest-list-v0.0.0
podman manifest rm ${REPOSITORY}:docker-v2-manifest-list-v0.0.1
podman manifest rm ${REPOSITORY}:docker-v2-manifest-list-v0.0.2

# oci v1 image index
echo -e "FROM scratch\nLABEL org.opencontainers.image.version=0.0.0\n" | podman build -f - --no-cache --timestamp=0 --platform=linux/amd64 --manifest ${REPOSITORY}:oci-v1-image-index-v0.0.0
echo -e "FROM scratch\nLABEL org.opencontainers.image.version=0.0.1\nLABEL org.opencontainers.image.source=https://github.com/airfocusio/git-ops-update\nLABEL org.opencontainers.image.revision=99a7aaa2eff5aff7c77d9caf0901454fedf6bf00\n" | podman build -f - --no-cache --timestamp=0 --platform=linux/amd64 --manifest ${REPOSITORY}:oci-v1-image-index-v0.0.1
echo -e "FROM scratch\nLABEL org.opencontainers.image.version=0.0.2\nLABEL org.opencontainers.image.source=https://github.com/airfocusio/git-ops-update\nLABEL org.opencontainers.image.revision=7f4304f54fd1f89aecc15ec3e70975e5bacfeb68\n" | podman build -f - --no-cache --timestamp=0 --platform=linux/amd64 --manifest ${REPOSITORY}:oci-v1-image-index-v0.0.2
podman manifest push --format=oci --all ${REPOSITORY}:oci-v1-image-index-v0.0.0 docker://${REPOSITORY}:oci-v1-image-index-v0.0.0
podman manifest push --format=oci --all ${REPOSITORY}:oci-v1-image-index-v0.0.1 docker://${REPOSITORY}:oci-v1-image-index-v0.0.1
podman manifest push --format=oci --all ${REPOSITORY}:oci-v1-image-index-v0.0.2 docker://${REPOSITORY}:oci-v1-image-index-v0.0.2
podman manifest rm ${REPOSITORY}:oci-v1-image-index-v0.0.0
podman manifest rm ${REPOSITORY}:oci-v1-image-index-v0.0.1
podman manifest rm ${REPOSITORY}:oci-v1-image-index-v0.0.2
