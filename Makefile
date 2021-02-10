.PHONY: all
all: lint license build unit-test e2e-test doc-gen

.PHONY: lint
lint:
	golangci-lint run

.PHONY: license
license:
	addlicense -c "Tetrate" src/ cmd/ e2e/

.PHONY: build
build:
	go build .

.PHONY: unit-test
unit-test:
	go test -v ./src/... ./cmd/...

.PHONY: e2e-test
e2e-test:
	go test -v ./e2e/...

.PHONY: doc-gen
doc-gen:
	go run doc/gen.go
