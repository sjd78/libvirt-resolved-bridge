APP_NAME := libvirt-resolved-bridge
VERSION := 0.1.0
DIST_DIR := $(APP_NAME)-$(VERSION)

.PHONY: archive
archive:
	mkdir -p $(DIST_DIR)/src
	cp src/*.go $(DIST_DIR)/src/
	cp go.mod go.sum $(APP_NAME).service README.md LICENSE $(DIST_DIR)/
	tar -czvf $(DIST_DIR).tar.gz $(DIST_DIR)
	rm -rf $(DIST_DIR)

.PHONY: clean
clean:
	rm -f $(APP_NAME)
	rm -f *.tar.gz
