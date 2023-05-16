DOCS_BASE_BRANCH=unset
MINOR_BRANCH=$(shell echo $(CLEAN_VERSION) | sed 's/\([0-9]*\.[0-9]*\.\)[0-9]*/\1x/')
BUMPED_MINOR_BRANCH=$(shell echo $(BUMPED_CLEAN_VERSION) | sed 's/\([0-9]*\.[0-9]*\.\)[0-9]*/\1x/')
ifeq ($(BUMP),auto)
DOCS_BASE_BRANCH=master
else ifeq ($(BUMP), major)
DOCS_BASE_BRANCH=master
else ifeq ($(BUMP), minor)
DOCS_BASE_BRANCH=master
else ifeq ($(BUMP), patch)
DOCS_BASE_BRANCH=$(BUMPED_MINOR_BRANCH)
endif

OLDEST_BRANCH_CCLOUD=1.20.1-post
OLDEST_BRANCH_CONFLUENT=1.16.3-post

NEXT_MINOR_VERSION = $(word 1,$(split_version)).$(shell expr $(word 2,$(split_version)) + 1).0
SHORT_NEXT_MINOR_VERSION = $(word 1,$(split_version)).$(shell expr $(word 2,$(split_version)) + 1)
CURRENT_SHORT_MINOR_VERSION = $(word 1,$(split_version)).$(word 2,$(split_version))
NEXT_PATCH_VERSION = $(word 1,$(split_version)).$(word 2,$(split_version)).$(shell expr $(word 3,$(split_version)) + 1)

.PHONY: clone-docs-repos
clone-docs-repos:
	$(eval TMP_BASE=$(shell mktemp -d))
	$(eval CCLOUD_DOCS_DIR=$(TMP_BASE)/docs-ccloud-cli)
	$(eval CONFLUENT_DOCS_DIR=$(TMP_BASE)/docs-confluent-cli)
	git clone git@github.com:confluentinc/docs-ccloud-cli.git $(CCLOUD_DOCS_DIR)
	git clone git@github.com:confluentinc/docs-confluent-cli.git $(CONFLUENT_DOCS_DIR)
	for repo in $(CCLOUD_DOCS_DIR) $(CONFLUENT_DOCS_DIR) ; do \
		cd $${repo} && \
		git fetch && \
		git checkout $(DOCS_BASE_BRANCH) && \
		cd .. ; \
	done

.PHONY: docs
docs: clean-docs
	@go run -ldflags '-X main.cliName=ccloud' cmd/docs/main.go
	@go run -ldflags '-X main.cliName=confluent' cmd/docs/main.go

.PHONY: publish-docs
publish-docs: docs clone-docs-repos
	echo -n "Publish ccloud docs? (y/n) "; read line; \
	if [ $$line = "y" ] || [ $$line = "Y" ]; then make publish-docs-internal REPO_DIR=$(CCLOUD_DOCS_DIR) CLI_NAME=ccloud; fi; \
	echo -n "Publish confluent docs? (y/n) "; read line; \
	if [ $$line = "y" ] || [ $$line = "Y" ]; then make publish-docs-internal REPO_DIR=$(CONFLUENT_DOCS_DIR) CLI_NAME=confluent; fi;

.PHONY: publish-docs-internal
publish-docs-internal:
	@cd $(REPO_DIR); \
	git checkout -b $(CLI_NAME)-cli-$(VERSION) origin/$(DOCS_BASE_BRANCH) || exit 1; \
	rm -rf command-reference; \
	cp -R $(GOPATH)/src/github.com/confluentinc/cli/docs/$(CLI_NAME) command-reference; \
	[ ! -f "command-reference/kafka/topic/ccloud_kafka_topic_consume.rst" ] || sed -i '' 's/default "confluent_cli_consumer_[^"]*"/default "confluent_cli_consumer_<uuid>"/' command-reference/kafka/topic/ccloud_kafka_topic_consume.rst || exit 1; \
	git add . || exit 1; \
	git diff --cached --exit-code > /dev/null && echo "nothing to update for docs" && exit 0; \
	git commit -m "[ci skip] chore: update $(CLI_NAME) CLI docs for $(VERSION)" || exit 1; \
	git push origin $(CLI_NAME)-cli-$(VERSION) || exit 1; \
	hub pull-request -b $(DOCS_BASE_BRANCH) -m "chore: update $(CLI_NAME) CLI docs for $(VERSION)" || exit 1

.PHONY: clean-docs
clean-docs:
	@rm -rf docs/

# NB: This should be getting run after a version release has happened.
# So $(VERSION) is the version that was just released, and $(BUMPED_VERSION)
# would be the next minor release (something in the future that doesn't exist yet).
# NB2: If a patch release just happened, $(DOCS_BASE_BRANCH) will still be accurate.
# Warning: BUMP must be set to patch if you are releasing docs for a patch release that was just done
.PHONY: release-docs
release-docs: clone-docs-repos cut-docs-branches update-settings-and-conf

.PHONY: cut-docs-branches
cut-docs-branches:
	if [[ "$(BUMP)" == "patch" ]]; then \
		for repo in $(CONFLUENT_DOCS_DIR) ; do \
			cd $${repo} && \
			git fetch && \
			git checkout $(MINOR_BRANCH) && \
			git checkout -b $(CLEAN_VERSION)-post && \
			git push -u origin $(CLEAN_VERSION)-post && \
			cd .. ; \
		done ; \
	else \
		for repo in $(CONFLUENT_DOCS_DIR) ; do \
			cd $${repo} && \
			git fetch && \
			git checkout $(DOCS_BASE_BRANCH) && \
			git checkout -b $(MINOR_BRANCH) && \
			git push -u origin $(MINOR_BRANCH) && \
			git checkout -b $(CLEAN_VERSION)-post && \
			git push -u origin $(CLEAN_VERSION)-post && \
			cd .. ; \
		done; \
	fi

# NB: If BUMP is patch, we don't have to update master
.PHONY: update-settings-and-conf
update-settings-and-conf:
	if [[ "$(BUMP)" != "patch" ]]; then \
		for repo in $(CONFLUENT_DOCS_DIR) ; do \
			cd $${repo} && \
			git fetch && \
			git checkout master && \
			sed -i '' 's/export RELEASE_VERSION=.*/export RELEASE_VERSION=$(NEXT_MINOR_VERSION)-SNAPSHOT/g' settings.sh && \
			sed -i '' "s/^version = '.*'/version = \'$(SHORT_NEXT_MINOR_VERSION)\'/g" conf.py && \
			sed -i '' "s/^release = '.*'/release = \'$(NEXT_MINOR_VERSION)-SNAPSHOT\'/g" conf.py && \
			git commit -am "[ci skip] chore: update settings.sh and conf.py due to $(CLEAN_VERSION) release" && \
			git push && \
			cd .. ; \
		done ; \
	fi
	for repo in $(CONFLUENT_DOCS_DIR) ; do \
		cd $${repo} && \
		git fetch && \
		git checkout $(MINOR_BRANCH) && \
		sed -i '' 's/export RELEASE_VERSION=.*/export RELEASE_VERSION=$(NEXT_PATCH_VERSION)-SNAPSHOT/g' settings.sh && \
		sed -i '' "s/^version = '.*'/version = \'$(CURRENT_SHORT_MINOR_VERSION)\'/g" conf.py && \
		sed -i '' "s/^release = '.*'/release = \'$(NEXT_PATCH_VERSION)-SNAPSHOT\'/g" conf.py && \
		git commit -am "[ci skip] chore: update settings.sh and conf.py due to $(CLEAN_VERSION) release" && \
		git push && \
		git checkout $(CLEAN_VERSION)-post && \
		sed -i '' 's/export RELEASE_VERSION=.*/export RELEASE_VERSION=$(CLEAN_VERSION)/g' settings.sh && \
		sed -i '' "s/^version = '.*'/version = \'$(CURRENT_SHORT_MINOR_VERSION)\'/g" conf.py && \
		sed -i '' "s/^release = '.*'/release = \'$(CLEAN_VERSION)\'/g" conf.py && \
		git commit -am "[ci skip] chore: update settings.sh and conf.py due to $(CLEAN_VERSION) release" && \
		git push && \
		cd .. ; \
	done
