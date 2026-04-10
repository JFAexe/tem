default: install

@lint:
  golangci-lint run ./...

@build:
  just _go build

@install:
  just _go install

[env('CGO_ENABLED', '0')]
@_go cmd:
  go {{ cmd }} -trimpath -ldflags "-s -w -buildid=" ./cmd/tem
