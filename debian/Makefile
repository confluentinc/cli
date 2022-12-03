SHELL=/bin/bash

ifndef VERSION
	VERSION=$(CLI_VERSION)
endif

export PACKAGE_TITLE=cli
export FULL_PACKAGE_TITLE=confluent-$(PACKAGE_TITLE)
export PACKAGE_NAME=$(FULL_PACKAGE_TITLE)-$(VERSION)

# Defaults that are likely to vary by platform. These are cleanly separated so
# it should be easy to maintain altered values on platform-specific branches
# when the values aren't overridden by the script invoking the Makefile

APPLY_PATCHES?=yes

# DESTDIR may be overridden by e.g. debian packaging
ifeq ($(DESTDIR),)
DESTDIR=$(CURDIR)/BUILD/
endif

ifeq ($(PACKAGE_TYPE),archive)
PREFIX=$(PACKAGE_NAME)
SYSCONFDIR=$(PREFIX)/etc/$(PACKAGE_TITLE)
else
PREFIX=/usr
SYSCONFDIR=/etc/$(PACKAGE_TITLE)
endif

all: install

archive: install
	rm -f $(CURDIR)/$(PACKAGE_NAME).tar.gz && cd $(DESTDIR) && tar -czf $(CURDIR)/$(PACKAGE_NAME).tar.gz $(PREFIX)
	rm -f $(CURDIR)/$(PACKAGE_NAME).zip && cd $(DESTDIR) && zip -r $(CURDIR)/$(PACKAGE_NAME).zip $(PREFIX)

apply-patches: $(wildcard debian/patches/*)
ifeq ($(APPLY_PATCHES),yes)
	git reset --hard HEAD
	cat debian/patches/series | xargs -iPATCH bash -c 'patch -p1 < debian/patches/PATCH'
endif

BINPATH=$(PREFIX)/bin
LIBPATH=$(PREFIX)/libexec/$(PACKAGE_TITLE)
DOCPATH=$(PREFIX)/share/doc/$(PACKAGE_TITLE)

# Install a single binary, specified with $CLI_VERSION, $OS and $ARCH
install: apply-patches
	rm -rf $(DESTDIR)$(PREFIX)
	mkdir -p $(DESTDIR)$(PREFIX)

	mkdir -p $(DESTDIR)$(BINPATH)
	mkdir -p $(DESTDIR)$(DOCPATH)
	mkdir -p $(DESTDIR)$(SYSCONFDIR)

	curl -f -s https://s3-us-west-2.amazonaws.com/confluent.cloud/confluent-cli/binaries/$(CLI_VERSION)/confluent_$(CLI_VERSION)_linux_amd64 -o $(DESTDIR)$(BINPATH)/confluent
	chmod 755 $(DESTDIR)$(BINPATH)/confluent

	cp LICENSE $(DESTDIR)$(DOCPATH)/COPYRIGHT
	$(DESTDIR)$(BINPATH)/confluent --version | awk -F' ' '{ print $3 }' > $(DESTDIR)$(DOCPATH)/version.txt

# Install all OS/arch combinations and bundle with packaging/confluent.sh for OS/arch-agnostic TAR/ZIP archives
install-all: apply-patches
	rm -rf $(DESTDIR)$(PREFIX)
	mkdir -p $(DESTDIR)$(PREFIX)

	mkdir -p $(DESTDIR)$(BINPATH)
	mkdir -p $(DESTDIR)$(LIBPATH)
	mkdir -p $(DESTDIR)$(DOCPATH)
	mkdir -p $(DESTDIR)$(SYSCONFDIR)

	cp packaging/confluent.sh $(DESTDIR)$(BINPATH)/confluent
	chmod 755 $(DESTDIR)$(BINPATH)/confluent

	cd $(DESTDIR)$(LIBPATH) ; \
	for dir in darwin_amd64 darwin_arm64 linux_amd64 linux_arm64 windows_amd64; do \
		mkdir -p $${dir} ; \
		ext=""; if [[ $${dir} =~ windows_.+ ]]; then ext=".exe"; fi ; \
		filepath=$${dir}/confluent$${ext} ; \
		curl -f -s https://s3-us-west-2.amazonaws.com/confluent.cloud/confluent-cli/binaries/$(CLI_VERSION)/confluent_$(CLI_VERSION)_$${dir}$${ext} -o $${filepath} ; \
		chmod 755 $${filepath} ; \
	done

	cp LICENSE $(DESTDIR)$(DOCPATH)/COPYRIGHT
	$(DESTDIR)$(BINPATH)/confluent --version | awk -F' ' '{ print $3 }' > $(DESTDIR)$(DOCPATH)/version.txt

	chown -R root:root $(DESTDIR)$(PREFIX)

clean:
	rm -rf $(CURDIR)/$(PACKAGE_NAME)*
	rm -rf $(FULL_PACKAGE_TITLE)-$(RPM_VERSION)*rpm
	rm -rf RPM_BUILDING

distclean: clean
ifneq ($(PACKAGE_TYPE),deb)
	git reset --hard HEAD
	git status --ignored --porcelain | cut -d ' ' -f 2 | xargs rm -rf
endif
