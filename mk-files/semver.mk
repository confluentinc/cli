_empty :=
_space := $(_empty) $(empty)

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
