VERSION = $(shell godzil show-version)
CURRENT_REVISION = $(shell git rev-parse --short HEAD)
BUILD_LDFLAGS = "-s -w -X github.com/Songmu/podbard.revision=$(CURRENT_REVISION)"
u := $(if $(update),-u)

.PHONY: deps
deps:
	go get ${u}
	go mod tidy

.PHONY: devel-deps
devel-deps:
	go install github.com/Songmu/godzil/cmd/godzil@latest
	go install github.com/tcnksm/ghr@latest

.PHONY: test
test:
	go test

.PHONY: build
build:
	go build -ldflags=$(BUILD_LDFLAGS) ./cmd/podbard

.PHONY: install
install:
	go install -ldflags=$(BUILD_LDFLAGS) ./cmd/podbard

.PHONY: release
release: devel-deps
	godzil release

CREDITS: go.sum deps devel-deps
	godzil credits -w

DIST_DIR = dist/v$(VERSION)
.PHONY: crossbuild
crossbuild: CREDITS
	rm -rf $(DIST_DIR)
	godzil crossbuild -pv=v$(VERSION) -build-ldflags=$(BUILD_LDFLAGS) \
      -os=linux,darwin -d=$(DIST_DIR) ./cmd/*
	cd $(DIST_DIR) && shasum -a 256 $$(find * -type f -maxdepth 0) > SHA256SUMS

.PHONY: upload
upload:
	ghr -body="$$(godzil changelog --latest -F markdown)" v$(VERSION) dist/v$(VERSION)


.PHONY: sync
sync:
	cp testdata/dev/index.md internal/cast/testdata/init/index.md
	cp testdata/dev/static/css/style.css internal/cast/testdata/init/static/css/style.css
	cp testdata/dev/static/css/style.css docs/ja/static/css/style.css
	cp -r testdata/dev/template/ internal/cast/testdata/init/template/
	cp -r testdata/dev/static/ internal/cast/testdata/init/static/
