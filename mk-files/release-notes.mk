.PHONY: release-notes-prep
release-notes-prep:
	@echo "Preparing Release Notes for $(BUMPED_VERSION) (Previous Release Version: v$(CLEAN_VERSION))"
	@echo
	@GO11MODULE=on go run -ldflags '-X main.releaseVersion=$(BUMPED_VERSION) -X main.prevVersion=v$(CLEAN_VERSION)' cmd/release-notes/prep/main.go
	$(print-release-notes-prep-next-steps)

define print-release-notes-prep-next-steps
	@echo "===================="
	@echo "NEXT STEPS"
	@echo "===================="
	@echo
	@echo "- Open './release-notes/prep' and fill in the content following the instructions."
	@echo
	@echo "- Once finished, run 'make publish-release-notes'."
	@echo
	@echo "===================="
endef

RELEASE_NOTES_BRANCH ?= cli-$(BUMPED_VERSION)-release-notes
.PHONY: publish-release-notes
publish-release-notes: clone-and-setup-docs-repos
	make build-release-notes CCLOUD_DOCS_DIR=$(CCLOUD_DOCS_DIR) CONFLUENT_DOCS_DIR=$(CONFLUENT_DOCS_DIR)
	make publish-release-notes-to-s3 CCLOUD_DOCS_DIR=$(CCLOUD_DOCS_DIR) CONFLUENT_DOCS_DIR=$(CONFLUENT_DOCS_DIR)
	make publish-release-notes-to-docs-repos CCLOUD_DOCS_DIR=$(CCLOUD_DOCS_DIR) CONFLUENT_DOCS_DIR=$(CONFLUENT_DOCS_DIR)
	rm -rf $(TMP_BASE)
	$(print-publish-release-notes-next-steps)

.PHONY: clone-and-setup-docs-repos
clone-and-setup-docs-repos:
	$(eval TMP_BASE=$(shell mktemp -d))
	$(eval CCLOUD_DOCS_DIR=$(TMP_BASE)/docs-ccloud-cli)
	$(eval CONFLUENT_DOCS_DIR=$(TMP_BASE)/docs-confluent-cli)
	git clone git@github.com:confluentinc/docs-ccloud-cli.git $(CCLOUD_DOCS_DIR)
	git clone git@github.com:confluentinc/docs-confluent-cli.git $(CONFLUENT_DOCS_DIR)
	for repo in $(CCLOUD_DOCS_DIR) $(CONFLUENT_DOCS_DIR) ; do \
		cd $${repo} && \
		git fetch && \
		git checkout $(DOCS_BASE_BRANCH) && \
		git checkout -b $(RELEASE_NOTES_BRANCH) && \
		cd .. ; \
	done

.PHONY: build-release-notes
build-release-notes:
	@echo Previous Release Version: v$(CLEAN_VERSION)
	@GO11MODULE=on go run -ldflags '-X main.releaseVersion=$(BUMPED_VERSION) -X main.ccloudReleaseNotesPath=$(CCLOUD_DOCS_DIR) -X main.confluentReleaseNotesPath=$(CONFLUENT_DOCS_DIR)' cmd/release-notes/release/main.go

.PHONY: publish-release-notes-to-docs-repos
publish-release-notes-to-docs-repos:
	cp release-notes/ccloud/release-notes.rst $(CCLOUD_DOCS_DIR)
	cp release-notes/confluent/release-notes.rst $(CONFLUENT_DOCS_DIR)
	$(warning SUBMITTING PRs to docs repos)
	for repo in $(CCLOUD_DOCS_DIR) $(CONFLUENT_DOCS_DIR); do \
		cd $${repo} || exit 1; \
		git add . || exit 1; \
		git diff --cached --exit-code > /dev/null && echo "nothing to update" && exit 0; \
		git commit -m "New release notes for $(BUMPED_VERSION)" || exit 1; \
		git push origin $(RELEASE_NOTES_BRANCH) || exit 1; \
		hub pull-request -b $(DOCS_BASE_BRANCH) -m "New release notes for $(BUMPED_VERSION)" ; \
	done

.PHONY: publish-release-notes-to-s3
publish-release-notes-to-s3:
	aws s3 cp release-notes/ccloud/latest-release.rst $(S3_BUCKET_PATH)/ccloud-cli/release-notes/$(BUMPED_VERSION:v%=%)/release-notes.rst --acl public-read; \
    aws s3 cp release-notes/confluent/latest-release.rst $(S3_BUCKET_PATH)/confluent-cli/release-notes/$(BUMPED_VERSION:v%=%)/release-notes.rst --acl public-read

define print-publish-release-notes-next-steps
	@echo
	@echo
	@echo "===================="
	@echo "NEXT STEPS"
	@echo "===================="
	@echo
	@echo "- Find PRs named 'New release notes for $(BUMPED_VERSION)' in confluentinc/docs-ccloud-cli and confluentinc/docs-confluent-cli and merge them both after getting approvals."
	@echo
	@echo "- Check release notes file in s3 confluent.cloud/ccloud-cli/release-notes/$(BUMPED_VERSION:v%=%)/"
	@echo
	@echo "- Run 'make clean-release-notes' to clean up your local repo"
	@echo
	@echo "- Once the release notes are ready, it's time to release the CLI!"
	@echo
	@echo "===================="
endef

.PHONY: clean-release-notes
clean-release-notes:
	-rm release-notes/prep
	-rm release-notes/ccloud/release-notes.rst
	-rm release-notes/confluent/release-notes.rst
	-rm release-notes/ccloud/latest-release.rst
	-rm release-notes/confluent/latest-release.rst
