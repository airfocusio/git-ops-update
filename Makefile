.PHONY: *

run:
	go run . --dir=example --dry --verbose

test:
	go test -v ./...

test-watch:
	watch -n1 go test -v ./...

test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out

build:
	goreleaser release --rm-dist --skip-publish --snapshot

release:
	goreleaser release --rm-dist


generate-test-images: REPOSITORY=ghcr.io/airfocusio/git-ops-update-test
generate-test-images:
	# docker v2 manifest
	echo "FROM scratch\nLABEL version=0.0.1\nLABEL foo=bar" | podman build -f - --format=docker --timestamp=0 --no-cache --tag $(REPOSITORY):docker-v2-manifest-v0.0.1
	echo "FROM scratch\nLABEL version=0.0.2\nLABEL foo=bar" | podman build -f - --format=docker --timestamp=0 --no-cache --tag $(REPOSITORY):docker-v2-manifest-v0.0.2
	podman push $(REPOSITORY):docker-v2-manifest-v0.0.1
	podman push $(REPOSITORY):docker-v2-manifest-v0.0.2
	podman image rm $(REPOSITORY):docker-v2-manifest-v0.0.1
	podman image rm $(REPOSITORY):docker-v2-manifest-v0.0.2

	# oci v1 manifest
	echo "FROM scratch\nLABEL version=0.0.1\nLABEL foo=bar" | podman build -f - --format=oci --timestamp=0 --no-cache --tag $(REPOSITORY):oci-v1-manifest-v0.0.1
	echo "FROM scratch\nLABEL version=0.0.2\nLABEL foo=bar" | podman build -f - --format=oci --timestamp=0 --no-cache --tag $(REPOSITORY):oci-v1-manifest-v0.0.2
	podman push $(REPOSITORY):oci-v1-manifest-v0.0.1
	podman push $(REPOSITORY):oci-v1-manifest-v0.0.2
	podman image rm $(REPOSITORY):oci-v1-manifest-v0.0.1
	podman image rm $(REPOSITORY):oci-v1-manifest-v0.0.2

	# docker v2 manifest list
	echo "FROM scratch\nLABEL version=0.0.1\nLABEL foo=bar" | podman build -f - --no-cache --timestamp=0 --platform=linux/amd64 --manifest $(REPOSITORY):docker-v2-manifest-list-v0.0.1
	echo "FROM scratch\nLABEL version=0.0.2\nLABEL foo=bar" | podman build -f - --no-cache --timestamp=0 --platform=linux/amd64 --manifest $(REPOSITORY):docker-v2-manifest-list-v0.0.2
	podman manifest push --format=v2s2 --all $(REPOSITORY):docker-v2-manifest-list-v0.0.1 docker://$(REPOSITORY):docker-v2-manifest-list-v0.0.1
	podman manifest push --format=v2s2 --all $(REPOSITORY):docker-v2-manifest-list-v0.0.2 docker://$(REPOSITORY):docker-v2-manifest-list-v0.0.2
	podman manifest rm $(REPOSITORY):docker-v2-manifest-list-v0.0.1
	podman manifest rm $(REPOSITORY):docker-v2-manifest-list-v0.0.2

	# oci v1 image index
	echo "FROM scratch\nLABEL version=0.0.1\nLABEL foo=bar" | podman build -f - --no-cache --timestamp=0 --platform=linux/amd64 --manifest $(REPOSITORY):oci-v1-image-index-v0.0.1
	echo "FROM scratch\nLABEL version=0.0.2\nLABEL foo=bar" | podman build -f - --no-cache --timestamp=0 --platform=linux/amd64 --manifest $(REPOSITORY):oci-v1-image-index-v0.0.2
	podman manifest push --format=oci --all $(REPOSITORY):oci-v1-image-index-v0.0.1 docker://$(REPOSITORY):oci-v1-image-index-v0.0.1
	podman manifest push --format=oci --all $(REPOSITORY):oci-v1-image-index-v0.0.2 docker://$(REPOSITORY):oci-v1-image-index-v0.0.2
	podman manifest rm $(REPOSITORY):oci-v1-image-index-v0.0.1
	podman manifest rm $(REPOSITORY):oci-v1-image-index-v0.0.2
