.PHONY: unrelease
unrelease: unrelease-warn
	make unrelease-s3
ifneq (true, $(RELEASE_TEST))
	$(warning Unreleasing on master)
	git checkout master
	git pull
else
	$(warning Unrelease test run)
endif
	git diff-index --quiet HEAD # ensures git status is clean
	git tag -d v$(CLEAN_VERSION) # delete local tag
	git push --delete origin v$(CLEAN_VERSION) # delete remote tag
	git reset --hard HEAD~1 # warning: assumes "chore" version bump was last commit
	git push origin HEAD --force
	make restore-latest-archives

.PHONY: unrelease-warn
unrelease-warn:
	@echo "Latest tag:"
	@git describe --tags `git rev-list --tags --max-count=1`
	@echo "Latest commits:"
	@git --no-pager log --decorate=short --pretty=oneline -n10
	@echo -n "Warning: Ensure a git version bump (new commit and new tag) has occurred before continuing, else you will remove the prior version. Continue? (y/n): "
	@read line; if [ $$line = "n" ] || [ $$line = "N" ]; then echo aborting; exit 1; fi

.PHONY: unrelease-s3
unrelease-s3:
	@echo "If you are going to reattempt the release again without the need to edit the release notes, there is no need to delete the release notes from S3."
	@echo -n "Do you want to delete the release notes from S3? (y/n): "
	@read line; if [ $$line = "y" ] || [ $$line = "Y" ]; then make delete-binaries-archives-and-release-notes; else make delete-binaries-and-archives; fi

.PHONY: delete-binaries-and-archives
delete-binaries-and-archives:
	$(caasenv-authenticate); \
	$(delete-binaries); \
	$(delete-archives)

.PHONY: delete-binaries-archives-and-release-notes
delete-binaries-archives-and-release-notes:
	$(caasenv-authenticate); \
	$(delete-binaries); \
	$(delete-archives); \
	$(delete-release-notes)

define delete-binaries
	aws s3 rm $(S3_BUCKET_PATH)/ccloud-cli/binaries/$(CLEAN_VERSION) --recursive; \
	aws s3 rm $(S3_BUCKET_PATH)/confluent-cli/binaries/$(CLEAN_VERSION) --recursive
endef

define delete-archives
	aws s3 rm $(S3_BUCKET_PATH)/ccloud-cli/archives/$(CLEAN_VERSION) --recursive; \
	aws s3 rm $(S3_BUCKET_PATH)/confluent-cli/archives/$(CLEAN_VERSION) --recursive
endef

define delete-release-notes
	aws s3 rm $(S3_BUCKET_PATH)/ccloud-cli/release-notes/$(CLEAN_VERSION) --recursive; \
	aws s3 rm $(S3_BUCKET_PATH)/confluent-cli/release-notes/$(CLEAN_VERSION) --recursive
endef

.PHONY: restore-latest-archives
restore-latest-archives: restore-latest-archives-warn
	make copy-archives-to-latest
	@echo "Verifying latest archives with: make test-installers"
	make test-installers

.PHONY: restore-latest-archives-warn
restore-latest-archives-warn:
	@echo -n "Warning: Overriding archives in the latest folder with archives from version v$(CLEAN_VERSION). Continue? (y/n): "
	@read line; if [ $$line = "n" ] || [ $$line = "N" ]; then echo aborting; exit 1; fi
