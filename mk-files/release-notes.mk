.PHONY: release-notes
release-notes:
	$(eval TMP_BASE=$(shell mktemp -d))
	$(eval CONFLUENT_DOCS_DIR=$(TMP_BASE)/docs-confluent-cli)

	git clone git@github.com:confluentinc/docs-confluent-cli.git $(CONFLUENT_DOCS_DIR) && \
	go run -ldflags '-X main.releaseNotesPath=$(CONFLUENT_DOCS_DIR)' cmd/releasenotes/main.go && \
	bump=$$(cat release-notes/bump.txt) && \
	version=$$(cat release-notes/version.txt) && \
	cd $(CONFLUENT_DOCS_DIR) && \
	if [ "$${bump}" = "patch" ]; then \
		git checkout $$(echo $${version} | sed $(STAGING_BRANCH_REGEX)); \
	fi && \
	git checkout -b cli-v$${version}-release-notes && \
	cd - && \
	CONFLUENT_DOCS_DIR=$(CONFLUENT_DOCS_DIR) make publish-release-notes-to-docs-repo && \
	rm -rf $(TMP_BASE)

.PHONY: publish-release-notes-to-docs-repo
publish-release-notes-to-docs-repo:
	bump=$$(cat release-notes/bump.txt) && \
	version=$$(cat release-notes/version.txt) && \
	cp release-notes/release-notes.rst $(CONFLUENT_DOCS_DIR) && \
	cd $(CONFLUENT_DOCS_DIR) && \
	git commit -am "New release notes for v$${version}" && \
	git push -u origin cli-v$${version}-release-notes && \
	base="master" && \
	if [ "$${bump}" = "patch" ]; then \
		base=$$(echo $${version} | sed $(STAGING_BRANCH_REGEX)); \
	fi && \
	gh pr create -B $${base} --title "New release notes for v$${version}" --body ""

.PHONY: publish-release-notes-to-s3
publish-release-notes-to-s3:
	$(aws-authenticate); \
    aws s3 cp release-notes/latest-release.rst $(S3_BUCKET_PATH)/confluent-cli/release-notes/$$(cat release-notes/version.txt)/release-notes.rst --acl public-read

.PHONY: clean-release-notes
clean-release-notes:
	rm -r release-notes/
