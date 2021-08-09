.PHONY: all
all: lint license build test e2e-test doc-gen

.PHONY: lint
lint:
	golangci-lint run

.PHONY: license
license:
	addlicense -c "Tetrate" internal/ cmd/ e2e/

.PHONY: build
build:
	go build .

.PHONY: test
test:
	go test -v ./internal/... ./cmd/...

.PHONY: e2e-test
e2e-test:
	go test -v -count=1 ./e2e/...

.PHONY: doc-gen
doc-gen:
	go run doc/gen.go
