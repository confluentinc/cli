.PHONY: unrelease
unrelease: unrelease-warn
	make unrelease-s3
	git checkout master
	git pull
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
restore-latest-archives:
	make echo-stuff
	$(eval TEMP_FILE=$(shell mktemp -d))
	cd $(TEMP_FILE) ; \
	LATEST_VERSION=$(CLEAN_VERSION)
	for binary in ccloud confluent; do \
		aws s3 cp $(S3_BUCKET_PATH)/$${binary}-cli/archives/$(CLEAN_VERSION) $(TEMP_FILE)/$${binary}-cli --recursive ; \
		cd $(TEMP_FILE)/$${binary}-cli ; \
		for fname in $${binary}_v$(CLEAN_VERSION)_*; do \
			newname=`echo "$$fname" | sed 's/_v$(CLEAN_VERSION)/_latest/g'`; \
			mv $$fname $$newname; \
		done ; \
		rm *checksums.txt; \
		$(SHASUM) $${binary}_latest_* > $${binary}_latest_checksums.txt ; \
		aws s3 cp ./ $(S3_BUCKET_PATH)/$${binary}-cli/archives/latest --recursive ; \
	done

.PHONY: echo-stuff
echo-stuff:
	echo $(VERSION)
	echo $(VERSION_NO_V)
	echo $(CLEAN_VERSION)
	echo $(BUMPED_CLEAN_VERSION)
