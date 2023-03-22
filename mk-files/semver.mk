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

.PHONY: show-version
## Show version variables
show-version:
	@echo version: $(VERSION)
	@echo version no v: $(VERSION_NO_V)
	@echo clean version: $(CLEAN_VERSION)
	@echo version post append: $(VERSION_POST)

.PHONY: commit-release
commit-release:
	echo "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" width=\"94\" height=\"20\"><linearGradient id=\"b\" x2=\"0\" y2=\"100%\"><stop offset=\"0\" stop-color=\"#bbb\" stop-opacity=\".1\"/><stop offset=\"1\" stop-opacity=\".1\"/></linearGradient><clipPath id=\"a\"><rect width=\"94\" height=\"20\" rx=\"3\" fill=\"#fff\"/></clipPath><g clip-path=\"url(#a)\"><path fill=\"#555\" d=\"M0 0h49v20H0z\"/><path fill=\"#007ec6\" d=\"M49 0h45v20H49z\"/><path fill=\"url(#b)\" d=\"M0 0h94v20H0z\"/></g><g fill=\"#fff\" text-anchor=\"middle\" font-family=\"DejaVu Sans,Verdana,Geneva,sans-serif\" font-size=\"110\"><text x=\"255\" y=\"150\" fill=\"#010101\" fill-opacity=\".3\" transform=\"scale(.1)\" textLength=\"390\">release</text><text x=\"255\" y=\"140\" transform=\"scale(.1)\" textLength=\"390\">release</text><text x=\"705\" y=\"150\" fill=\"#010101\" fill-opacity=\".3\" transform=\"scale(.1)\" textLength=\"350\">v$$(cat release-notes/version.txt)</text><text x=\"705\" y=\"140\" transform=\"scale(.1)\" textLength=\"350\">v$$(cat release-notes/version.txt)</text></g> </svg>" > release.svg && \
	git add release.svg && \
	git commit -m "chore: $$(cat release-notes/bump.txt) version bump v$$(cat release-notes/version.txt) [ci skip]" && \
	$(call dry-run,git push)

.PHONY: tag-release
tag-release:
	# Delete tag from the remote in case it already exists
	git tag -d v$$(cat release-notes/version.txt) || true
	$(call dry-run,git push -d $(GIT_REMOTE_NAME) v$$(cat release-notes/version.txt)) || true

	# Add tag to the remote
	git tag v$$(cat release-notes/version.txt)
	$(call dry-run,git push $(GIT_REMOTE_NAME) v$$(cat release-notes/version.txt))
