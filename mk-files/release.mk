ARCHIVE_TYPES=darwin_amd64.tar.gz darwin_arm64.tar.gz linux_amd64.tar.gz linux_arm64.tar.gz alpine_amd64.tar.gz alpine_arm64.tar.gz windows_amd64.zip

# If you set up your laptop following https://github.com/confluentinc/cc-documentation/blob/master/Operations/Laptop%20Setup.md
# then assuming caas.sh lives here should be fine
define aws-authenticate
	$(call dry-run,export AWS_PROFILE=cc-production-1/prod-administrator)
endef


.PHONY: release
release: check-branch
	$(call print-boxed-message,"RELEASING TO STAGING FOLDER $(S3_STAG_PATH)")
	$(MAKE) release-to-stag
	$(call print-boxed-message,"PUBLISHING RELEASE NOTES TO S3 $(S3_BUCKET_PATH)")
	$(MAKE) publish-release-notes-to-s3
	$(call print-boxed-message,"PUBLISHING INSTALLER TO S3 $(S3_BUCKET_PATH)")
	$(MAKE) publish-installer
	$(call print-boxed-message,"RELEASING TO PROD FOLDER $(S3_BUCKET_PATH)")
	$(MAKE) release-to-prod
	$(call print-boxed-message,"UPDATING PACKAGING")
	$(MAKE) update-packaging
	$(call print-boxed-message,"UPDATING MUCKRAKE")
	$(MAKE) update-muckrake
	$(call print-boxed-message,"PUBLISHING DOCS")
	$(MAKE) publish-docs

.PHONY: check-branch
check-branch:
	if [ $(shell git rev-parse --abbrev-ref HEAD) != main ] ; then \
		echo -n "WARNING: Current branch \"$(shell git rev-parse --abbrev-ref HEAD)\" is not the default release branch \"main\"!  Do you want to proceed? (y/n): "; \
		read line; if [ $$line != "y" ] && [ $$line != "Y" ]; then echo "Release cancelled."; exit 0; fi; \
	fi

.PHONY: release-to-stag
release-to-stag:
	$(MAKE) gorelease
	$(MAKE) goreleaser-patches
	$(MAKE) copy-stag-archives-to-latest
	$(call print-boxed-message,"VERIFYING STAGING RELEASE CONTENT")
	$(MAKE) verify-stag
	$(call print-boxed-message,"STAGING RELEASE COMPLETED AND VERIFIED!")

.PHONY: release-to-prod
release-to-prod:
	$(aws-authenticate) && \
	$(call copy-stag-content-to-prod,archives,$(CLEAN_VERSION)); \
	$(call copy-stag-content-to-prod,binaries,$(CLEAN_VERSION)); \
	$(call copy-stag-content-to-prod,archives,latest); \
	$(call dry-run, aws s3 sync $(S3_DEB_RPM_STAG_PATH)/$(VERSION_NO_V)/deb $(S3_DEB_RPM_PROD_PATH)/deb); \
	$(call dry-run, aws s3 sync $(S3_DEB_RPM_STAG_PATH)/$(VERSION_NO_V)/rpm $(S3_DEB_RPM_PROD_PATH)/rpm); \
	$(call dry-run, s3-repo-utils -v website index --fake-index --prefix $(S3_DEB_RPM_PROD_PREFIX)/ $(S3_DEB_RPM_BUCKET_NAME))
	$(call print-boxed-message,"VERIFYING PROD RELEASE CONTENT")
	$(MAKE) verify-prod
	$(call print-boxed-message,"PROD RELEASE COMPLETED AND VERIFIED!")

define copy-stag-content-to-prod
	folder_path=confluent-cli/$1/$2; \
	echo "COPYING: $${folder_path}"; \
	$(call dry-run,aws s3 cp $(S3_STAG_PATH)/$${folder_path} $(S3_BUCKET_PATH)/$${folder_path} --recursive --acl public-read) || exit 1
endef

.PHONY: gorelease-linux-amd64
gorelease-linux-amd64:
	go install github.com/goreleaser/goreleaser@$(GORELEASER_VERSION) && \
	goreleaser release --clean --config .goreleaser-linux-amd64.yml

.PHONY: gorelease-linux-arm64
gorelease-linux-arm64:
ifneq (,$(findstring x86_64,$(shell uname -m)))
	go install github.com/goreleaser/goreleaser@$(GORELEASER_VERSION) && \
	CC=aarch64-linux-gnu-gcc CXX=aarch64-linux-gnu-g++ goreleaser release --clean --config .goreleaser-linux-arm64.yml
else
	go install github.com/goreleaser/goreleaser@$(GORELEASER_VERSION) && \
	goreleaser release --clean --config .goreleaser-linux-arm64.yml
endif

# This builds the Darwin, Windows and Linux binaries using goreleaser on the host computer. Goreleaser takes care of uploading the resulting binaries/archives/checksums to S3.
.PHONY: gorelease
gorelease:
	$(eval DIR=$(shell mktemp -d))
	$(eval CLI_RELEASE=$(DIR)/cli-release)

	$(eval token := $(shell (grep github.com ~/.netrc -A 2 | grep password || grep github.com ~/.netrc -A 2 | grep login) | head -1 | awk -F' ' '{ print $$2 }'))
	$(aws-authenticate) && \
	rm -rf prebuilt/ deb/ rpm/ && \
	mkdir prebuilt/ deb/ rpm/ && \
	scripts/build_linux.sh && \
	$(call dry-run,aws s3 sync deb $(S3_DEB_RPM_STAG_PATH)/$(VERSION_NO_V)/deb) && \
	$(call dry-run,aws s3 sync rpm $(S3_DEB_RPM_STAG_PATH)/$(VERSION_NO_V)/rpm) && \
	git clone git@github.com:confluentinc/cli-release.git $(CLI_RELEASE) && \
	go run $(CLI_RELEASE)/cmd/releasenotes/formatter/main.go $(CLI_RELEASE)/release-notes/$(VERSION_NO_V).json github > $(DIR)/release-notes.txt && \
	GORELEASER_KEY=$(GORELEASER_KEY) S3FOLDER=$(S3_STAG_FOLDER_NAME)/confluent-cli GITHUB_TOKEN=$(token) DRY_RUN=$(DRY_RUN) goreleaser release --clean --release-notes $(DIR)/release-notes.txt --timeout 60m && \
	sha256sum prebuilt/confluent-cli_$(VERSION_NO_V)*.deb prebuilt/confluent-cli-$(VERSION_NO_V)*.rpm >> dist/confluent_$(VERSION_NO_V)_checksums.txt && \
	$(call dry-run,gh release upload $(VERSION) prebuilt/*.deb prebuilt/*.rpm dist/confluent_$(VERSION_NO_V)_checksums.txt --clobber)

# Current goreleaser still has some shortcomings for the our use, and the target patches those issues
# As new goreleaser versions allow more customization, we may be able to reduce the work for this make target
.PHONY: goreleaser-patches
goreleaser-patches:
	$(MAKE) set-acls

# goreleaser does not yet support setting ACLs for cloud storage
# We have to set `public-read` manually by copying the file in place
# Dummy metadata is used as a hack because S3 does not allow copying files to the same place without any changes (--acl change doesn't count)
.PHONY: set-acls
set-acls:
	$(aws-authenticate) && \
	for file_type in archives binaries; do \
		folder_path=confluent-cli/$${file_type}/$(VERSION_NO_V); \
		echo "SETTING ACLS: $${folder_path}"; \
		$(call dry-run,aws s3 cp $(S3_STAG_PATH)/$${folder_path} $(S3_STAG_PATH)/$${folder_path} --acl public-read --metadata dummy=dummy --recursive) || exit 1; \
	done

# Update latest archives folder for staging
# Also used by unrelease to fix latest archives folder so have to be careful about the version variable used
# e.g. may be using this to restore while cli repo VERSION value is dirty, hence CLEAN_VERSION variable is used
.PHONY: copy-stag-archives-to-latest
copy-stag-archives-to-latest:
	$(call copy-archives-files-to-latest,$(S3_STAG_PATH),$(S3_STAG_PATH))
	$(call copy-archives-checksums-to-latest,$(S3_STAG_PATH),$(S3_STAG_PATH))

# first argument: S3 folder of archives we want to copy from
# second argument: S3 folder destination for latest archives
define copy-archives-files-to-latest
	$(aws-authenticate); \
	archives_folder=$1/confluent-cli/archives/$(CLEAN_VERSION); \
	latest_folder=$2/confluent-cli/archives/latest; \
	for suffix in $(ARCHIVE_TYPES); do \
		$(call dry-run,aws s3 cp $${archives_folder}/confluent_$(CLEAN_VERSION)_$${suffix} $${latest_folder}/confluent_latest_$${suffix} --acl public-read); \
	done
endef

# Copy archives checksum file then rename the filenames in the checksum by replacing VERSION to "latest"
# Then publish the checksum file to S3 latest folder
# first argument: S3 folder of archives we want to copy from
# second argument: S3 folder destination for latest archives
define copy-archives-checksums-to-latest
	$(eval DIR=$(shell mktemp -d))

	$(aws-authenticate); \
	version_checksums=confluent_$(CLEAN_VERSION)_checksums.txt; \
	latest_checksums=confluent_latest_checksums.txt; \
	cd $(DIR); \
	$(call dry-run,aws s3 cp $1/confluent-cli/archives/$(CLEAN_VERSION)/$${version_checksums} ./); \
	$(call dry-run,cat $${version_checksums} | grep "$(CLEAN_VERSION)" | sed 's/$(CLEAN_VERSION)/latest/' > $${latest_checksums}); \
	$(call dry-run,aws s3 cp $${latest_checksums} $2/confluent-cli/archives/latest/$${latest_checksums} --acl public-read)

	rm -rf $(DIR)
endef

.PHONY: download-licenses
download-licenses:
	go install github.com/google/go-licenses@v1.5.0 && \
	go-licenses save ./... --save_path legal/licenses/ --force || true && \
	chmod -R +w legal/licenses/

.PHONY: publish-installer
## Publish install scripts to S3. You MUST re-run this if/when you update any install script.
publish-installer:
	$(aws-authenticate) && \
	$(call dry-run,aws s3 cp install.sh $(S3_BUCKET_PATH)/confluent-cli/install.sh --acl public-read)

.PHONY: update-muckrake
update-muckrake:
	$(eval DIR=$(shell mktemp -d))
	$(eval CLI_RELEASE=$(DIR)/cli-release)
	$(eval MUCKRAKE=$(DIR)/muckrake)

	git clone git@github.com:confluentinc/cli-release.git $(CLI_RELEASE) && \
	cd $(CLI_RELEASE) && \
	version=$$(ls release-notes | $(SED) "s/.json$$//" | sort --version-sort | tail -1) && \
	git clone git@github.com:confluentinc/muckrake.git $(MUCKRAKE) && \
	cd $(MUCKRAKE) && \
	git fetch --all && \
	branch=bump-cli && \
	base=$$(git branch --remote --format "%(refname:short)" | sed -n "s|^origin/\([1-9][0-9]*\.[0-9][0-9]*\.x\)$$|\1|p" | tail -1) && \
	git checkout $$base && \
	git checkout $$branch || git checkout -b $$branch && \
	$(SED) -i "s|confluent-cli-.*=\$${confluent_s3}/confluent\.cloud/confluent-cli/archives/.*/confluent_.*_linux_amd64\.tar\.gz|confluent-cli-$${version}=\$${confluent_s3}/confluent.cloud/confluent-cli/archives/$${version}/confluent_$${version}_linux_amd64.tar.gz|" ducker/ducker && \
	$(SED) -i "s|VERSION = \".*\"|VERSION = \"$${version}\"|" muckrake/services/cli.py && \
	$(SED) -i "s|get_cli .*|get_cli $${version}|" vagrant/base-redhat.sh && \
	$(SED) -i "s|get_cli .*|get_cli $${version}|" vagrant/base-redhat8.sh && \
	$(SED) -i "s|get_cli .*|get_cli $${version}|" vagrant/base-redhat9.sh && \
	$(SED) -i "s|get_cli .*|get_cli $${version}|" vagrant/base-ubuntu.sh && \
	git commit -am "bump cli to v$${version}" && \
	$(call dry-run,git push -u origin $$branch) && \
	if gh pr view $$branch --json state --jq .state | grep "no pull requests found|MERGED"; then \
		$(call dry-run,gh pr create --base $${base} --title "Bump CLI to v$${version}" --body "") && \
		$(call dry-run,gh pr merge --squash --auto); \
	fi

	rm -rf $(DIR)

.PHONY: update-packaging
update-packaging:
	$(eval DIR=$(shell mktemp -d))
	$(eval CLI_RELEASE=$(DIR)/cli-release)
	$(eval PACKAGING=$(DIR)/packaging)

	git clone git@github.com:confluentinc/cli-release.git $(CLI_RELEASE) && \
	cd $(CLI_RELEASE) && \
	version=$$(ls release-notes | $(SED) "s/.json$$//" | sort --version-sort | tail -1) && \
	git clone git@github.com:confluentinc/packaging.git $(PACKAGING) && \
	cd $(PACKAGING) && \
	git fetch --all && \
	branch="bump-cli" && \
	base=$$(git branch --remote --format "%(refname:short)" | sed -n "s|^origin/\([1-9][0-9]*\.[0-9][0-9]*\.x\)$$|\1|p" | tail -1) && \
	git checkout $$base && \
	git checkout $$branch || git checkout -b $$branch && \
	$(SED) -i "s|cli_BRANCH=\".*\"|cli_BRANCH=\"$${version}\"|" settings.sh && \
	$(SED) -i "s|CLI_VERSION=.*|CLI_VERSION=$${version}|" release_testing/bin/smoke_test.sh && \
	git commit -am "bump cli to v$${version}" && \
	$(call dry-run,git push -u origin $$branch) && \
	if gh pr view $$branch --json state --jq .state | grep "no pull requests found|MERGED"; then \
		$(call dry-run,gh pr create --base $${base} --title "Bump CLI to v$${version}" --body "") && \
		$(call dry-run,gh pr merge --squash --auto); \
	fi

	rm -rf $(DIR)
