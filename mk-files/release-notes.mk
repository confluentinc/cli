.PHONY: publish-release-notes-to-s3
publish-release-notes-to-s3:
	$(eval DIR=$(shell mktemp -d))
	$(eval CLI_RELEASE=$(DIR)/cli-release)

	git clone git@github.com:confluentinc/cli-release.git $(CLI_RELEASE) && \
	version=$$(ls $(CLI_RELEASE)/release-notes | sed -e s/.json$$// | sort --version-sort | tail -1) && \
	go run cmd/releasenotes/main.go $(CLI_RELEASE)/release-notes/$${version}.json s3 > $(DIR)/release.txt && \
	$(aws-authenticate) && \
    $(call dry-run,aws s3 cp $(DIR)/release.txt $(S3_BUCKET_PATH)/confluent-cli/release-notes/$${version}/release-notes.rst --acl public-read)

	rm -rf $(DIR)
