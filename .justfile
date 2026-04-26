default: install

[doc('Run golangci-lint')]
@lint:
  golangci-lint run ./...

[doc('Run local goreleaser')]
@dist:
  goreleaser release --clean --snapshot

[doc('Update go.mod dependencies')]
@update:
  go get -u ./...

[doc('Build binary')]
@build:
  just _go build

[doc('Build and install binary')]
@install:
  just _go install

[env('CGO_ENABLED', '0')]
@_go cmd:
  go {{ cmd }} -trimpath -ldflags "-s -w -buildid=" ./cmd/tem
