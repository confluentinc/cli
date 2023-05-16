_empty :=
_space := $(_empty) $(empty)

# Gets added after the version
VERSION_POST ?=

VERSION ?= $(shell git rev-parse --is-inside-work-tree > /dev/null && git describe --tags --always --dirty)
ifneq (,$(findstring dirty,$(VERSION)))
VERSION := $(VERSION)-$(USER)
endif
CLEAN_VERSION ?= $(shell echo $(VERSION) | grep -Eo '([0-9]+\.){2}[0-9]+')
VERSION_NO_V := $(shell echo $(VERSION) | sed 's,^v,,' )

ifeq ($(CLEAN_VERSION),$(_empty))
CLEAN_VERSION := 0.0.0
endif

split_version := $(subst .,$(_space),$(CLEAN_VERSION))

.PHONY: tag-release
tag-release:
	$(eval DIR=$(shell mktemp -d))
	$(eval CLI_RELEASE=$(DIR)/cli-release)
	
	git clone git@github.com:confluentinc/cli-release.git $(CLI_RELEASE) && \
	version=$$(ls $(CLI_RELEASE)/release-notes | sed -e s/.json$$// | sort --version-sort | tail -1) && \
	git tag -d v$${version} || true && \
	$(call dry-run,git push -d origin v$${version}) || true && \
	git tag v$${version} && \
	$(call dry-run,git push origin v$${version})

	rm -rf $(DIR)
