LAST_TAG      = $(shell git describe --tags --abbrev=0 HEAD)
COMMIT        = $(shell git rev-parse --short HEAD)
FULL_COMMIT   = $(shell git rev-parse HEAD)
RELEASE_NOTES = $(shell git log ${LAST_TAG}..HEAD --oneline --decorate)
DATE          = $(shell date +%Y-%m-%d)

lint:
	golangci-lint run *.go

release_notes:
	@echo "${RELEASE_NOTES}"

before_build:
	go get github.com/mitchellh/gox

check_tag:
	test ! -z "${TAG}"

build: check_tag
	gox -verbose \
	    -output "dist/{{.Dir}}_{{.OS}}_{{.Arch}}" \
	    -osarch "linux/amd64 linux/arm64 linux/arm darwin/amd64 windows/amd64" \
	    -ldflags "-X main.buildCommit=${COMMIT} \
	              -X main.buildDate=${DATE} \
	              -X main.buildVersion=${LAST_TAG}" \
	    ./...

release: build
	@ghr -u wi1dcard -b "${RELEASE_NOTES}" -c "${FULL_COMMIT}" "${TAG}" dist/

docker_build: build
	docker build --build-arg ARCH=amd64 -t "wi1dcard/v2ray-exporter:${TAG}" .
	docker build --build-arg ARCH=arm64 -t "wi1dcard/v2ray-exporter:${TAG}-arm64" .
	docker build --build-arg ARCH=arm -t "wi1dcard/v2ray-exporter:${TAG}-arm" .
