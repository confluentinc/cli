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
	# Delete tag from the remote in case it already exists
	git tag -d v$$(cat release-notes/version.txt) || true
	$(call dry-run,git push -d origin v$$(cat release-notes/version.txt)) || true

	# Add tag to the remote
	git tag v$$(cat release-notes/version.txt)
	$(call dry-run,git push origin v$$(cat release-notes/version.txt))
