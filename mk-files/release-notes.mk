.PHONY: release-notes
release-notes:
	$(eval DIR=$(shell mktemp -d))
	$(eval DOCS_CONFLUENT_CLI=$(DIR)/docs-confluent-cli)

	git clone git@github.com:confluentinc/docs-confluent-cli.git $(DOCS_CONFLUENT_CLI) && \
	go run -ldflags '-X main.releaseNotesPath=$(DOCS_CONFLUENT_CLI)' cmd/releasenotes/main.go && \
	bump=$$(cat release-notes/bump.txt) && \
	version=$$(cat release-notes/version.txt) && \
	cd $(DOCS_CONFLUENT_CLI) && \
	if [ "$${bump}" = "patch" ]; then \
		git checkout $$(echo $${version} | sed $(STAGING_BRANCH_REGEX)); \
	fi && \
	git checkout -b cli-v$${version}-release-notes && \
	cd - && \
	cp release-notes/release-notes.rst $(DOCS_CONFLUENT_CLI) && \
	cd $(DOCS_CONFLUENT_CLI) && \
	git commit -am "New release notes for v$${version}" && \
	$(call dry-run,git push -u origin cli-v$${version}-release-notes) && \
	base="master" && \
	if [ "$${bump}" = "patch" ]; then \
		base=$$(echo $${version} | sed $(STAGING_BRANCH_REGEX)); \
	fi && \
	$(call dry-run,gh pr create -B $${base} --title "New release notes for v$${version}" --body "")

	rm -rf $(DIR)

.PHONY: publish-release-notes-to-s3
publish-release-notes-to-s3:
	$(aws-authenticate); \
    $(call dry-run,aws s3 cp release-notes/latest-release.rst $(S3_BUCKET_PATH)/confluent-cli/release-notes/$$(cat release-notes/version.txt)/release-notes.rst --acl public-read)
