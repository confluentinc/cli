.PHONY: unrelease-prod
unrelease-prod: unrelease-warn
	$(MAKE) delete-archives-and-binaries # needs to be run before version tag is reverted
	$(MAKE) delete-release-notes # needs to be run before version tag is reverted
	$(MAKE) reset-tag-and-commit
	$(MAKE) restore-latest-archives # needs to be run after version tag is reverted

.PHONY: unrelease-stag
unrelease-stag: unrelease-warn
	$(MAKE) delete-release-notes
	$(MAKE) reset-tag-and-commit
	$(MAKE) clean-staging-folder

.PHONY: reset-tag-and-commit
reset-tag-and-commit:
	git checkout main
	git pull
	git diff-index --quiet HEAD # ensures git status is clean
	git tag -d v$(CLEAN_VERSION) # delete local tag
	$(call dry-run,git push --delete origin v$(CLEAN_VERSION)) # delete remote tag

.PHONY: unrelease-warn
unrelease-warn:
	@echo "Latest tag:"
	git describe --tags `git rev-list --tags --max-count=1`
	@echo "Latest commits:"
	git --no-pager log --decorate=short --pretty=oneline -n10
	@echo -n "Warning: Ensure a git version bump (new commit and new tag) has occurred before continuing, else you will remove the prior version. Continue? (y/n): "
	read line; if [ $$line = "n" ] || [ $$line = "N" ]; then echo aborting; exit 1; fi

.PHONY: delete-archives-and-binaries
delete-archives-and-binaries:
	$(aws-authenticate); \
	$(call delete-release-folder,binaries); \
	$(call delete-release-folder,archives)

.PHONY: delete-release-notes
delete-release-notes:
	@echo -n "Do you want to delete the release notes from S3? (y/n): "
	read line; if [ $$line = "y" ] || [ $$line = "Y" ]; then $(aws-authenticate); $(call delete-release-folder,release-notes); fi

define delete-release-folder
	$(call dry-run,aws s3 rm $(S3_BUCKET_PATH)/confluent-cli/$1/$(CLEAN_VERSION) --recursive)
endef

.PHONY: restore-latest-archives
restore-latest-archives: restore-latest-archives-warn
	$(MAKE) copy-prod-archives-to-stag-latest
	$(aws-authenticate); \
	$(call copy-stag-content-to-prod,archives,latest)
	@echo "Verifying latest archives with: make test-installer"
	$(MAKE) test-installer

.PHONY: copy-prod-archives-to-stag-latest
copy-prod-archives-to-stag-latest:
	$(call copy-archives-files-to-latest,$(S3_BUCKET_PATH),$(S3_STAG_PATH))
	$(call copy-archives-checksums-to-latest,$(S3_BUCKET_PATH),$(S3_STAG_PATH))
	OVERRIDE_S3_FOLDER=$(S3_STAG_FOLDER_NAME) ARCHIVES_VERSION="" $(MAKE) test-installer

.PHONY: restore-latest-archives-warn
restore-latest-archives-warn:
	@echo -n "Warning: Overriding archives in the latest folder with archives from version v$(CLEAN_VERSION). Continue? (y/n): "
	read line; if [ $$line = "n" ] || [ $$line = "N" ]; then echo aborting; exit 1; fi

.PHONY: clean-staging-folder
clean-staging-folder:
	$(aws-authenticate); \
	$(call dry-run,aws s3 rm $(S3_STAG_PATH) --recursive)
