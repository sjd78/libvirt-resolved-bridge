APP_NAME := libvirt-resolved-bridge
VERSION := 0.1.0
DIST_DIR := $(APP_NAME)-$(VERSION)

# Build directories (can be overridden for COPR)
BUILD_DIR := build
SRPM_DIR ?= $(BUILD_DIR)/srpms
RPM_DIR := $(BUILD_DIR)/rpms

# Installation directories (can be overridden)
PREFIX ?= /usr
BINDIR ?= $(PREFIX)/bin
UNITDIR ?= $(PREFIX)/lib/systemd/system

# Go build flags
GO_LDFLAGS := -s -w
GO_BUILD_FLAGS := -trimpath -ldflags "$(GO_LDFLAGS)"

.PHONY: all
all: build

#
# Development targets
#
.PHONY: build
build:
	go build $(GO_BUILD_FLAGS) -o $(APP_NAME) ./src

.PHONY: test
test:
	go test -v ./src

.PHONY: test-coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./src
	go tool cover -html=coverage.out -o coverage.html

.PHONY: lint
lint:
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not installed"; exit 1; }
	golangci-lint run ./src/...

.PHONY: clean
clean:
	rm -f $(APP_NAME)
	rm -rf $(BUILD_DIR)
	rm -f *.tar.gz

#
# Installation targets
#
.PHONY: install
install: build
	install -D -m 0755 $(APP_NAME) $(DESTDIR)$(BINDIR)/$(APP_NAME)
	install -D -m 0644 $(APP_NAME).service $(DESTDIR)$(UNITDIR)/$(APP_NAME).service

.PHONY: uninstall
uninstall:
	rm -f $(DESTDIR)$(BINDIR)/$(APP_NAME)
	rm -f $(DESTDIR)$(UNITDIR)/$(APP_NAME).service

#
# Distribution tarball (source archive for RPM builds)
#
.PHONY: archive
archive:
	mkdir -p $(DIST_DIR)/src
	cp src/*.go $(DIST_DIR)/src/
	cp Makefile go.mod go.sum $(APP_NAME).service $(APP_NAME).spec README.md LICENSE $(DIST_DIR)/
	tar -czvf $(DIST_DIR).tar.gz $(DIST_DIR)
	rm -rf $(DIST_DIR)

#
# RPM building
#
.PHONY: srpm
srpm: archive
	mkdir -p $(SRPM_DIR)
	rpmbuild -bs \
		--define "_sourcedir $(CURDIR)" \
		--define "_srcrpmdir $(abspath $(SRPM_DIR))" \
		$(APP_NAME).spec

.PHONY: rpm
rpm: archive
	mkdir -p $(RPM_DIR)
	rpmbuild -bb \
		--define "_sourcedir $(CURDIR)" \
		--define "_rpmdir $(CURDIR)/$(RPM_DIR)" \
		$(APP_NAME).spec

.PHONY: rpm-local
rpm-local: srpm
	@command -v mock >/dev/null 2>&1 || { echo "mock not installed. Install with: sudo dnf install mock"; exit 1; }
	mock --rebuild $(SRPM_DIR)/$(APP_NAME)-$(VERSION)-*.src.rpm

.PHONY: help
help:
	@echo "$(APP_NAME) build targets:"
	@echo ""
	@echo "Development:"
	@echo "  build         - Build the binary"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  lint          - Run golangci-lint (if installed)"
	@echo "  clean         - Remove build artifacts"
	@echo ""
	@echo "Installation:"
	@echo "  install     - Install binary and service file"
	@echo "  uninstall   - Remove installed files"
	@echo ""
	@echo "Packaging:"
	@echo "  archive     - Create source tarball"
	@echo "  srpm        - Build source RPM"
	@echo "  rpm         - Build binary RPM (uses rpmbuild)"
	@echo "  rpm-local   - Build RPM using mock (isolated build)"
