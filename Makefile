.PHONY: build
build:
	go build -v -ldflags="-X main.SemVer=`git describe --tags --abbrev=0`"

.PHONY: release
release:
	-rm -rf dist/
	-rm ./gdbc
	goreleaser release
