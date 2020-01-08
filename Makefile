LAST_TAG  = $(shell git describe --tags --abbrev=0 HEAD)
COMMIT = $(shell git rev-parse --short HEAD)
FULL_COMMIT = $(shell git rev-parse HEAD)
RELEASE_NOTES = $(shell git log ${LAST_TAG}..HEAD --oneline --decorate)
DATE = $(shell date +%Y-%m-%d)

build:
	gox -verbose \
	    -output "dist/{{.Dir}}_{{.OS}}_{{.Arch}}" \
	    -osarch "linux/amd64 linux/arm64 linux/arm darwin/amd64 windows/amd64" \
	    -ldflags "-X main.buildCommit=${COMMIT} \
	              -X main.buildDate=${DATE} \
	              -X main.buildVersion=${LAST_TAG}" \
	    ./...

before_build:
	go get github.com/mitchellh/gox

lint:
	golangci-lint run *.go

release_notes:
	@echo "${RELEASE_NOTES}"

release: build
	test ! -z "${TAG}"
	@ghr -u wi1dcard -b "${RELEASE_NOTES}" -c "${FULL_COMMIT}" "${TAG}" dist/
