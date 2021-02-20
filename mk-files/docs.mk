DOCS_REPOS=docs-ccloud-cli docs-confluent-cli

DOCS_BASE_BRANCH=unset
ifeq ($(BUMP),auto)
DOCS_BASE_BRANCH=master
else ifeq ($(BUMP), major)
DOCS_BASE_BRANCH=master
else ifeq ($(BUMP), minor)
DOCS_BASE_BRANCH=master
else ifeq ($(BUMP), patch)
DOCS_BASE_BRANCH=$(shell echo $(BUMPED_CLEAN_VERSION) | sed 's/\([0-9]*\.[0-9]*\.\)[0-9]*/\1x/')
endif

.PHONY: cut-docs-branches
cut-docs-branches:
	echo "DOCS_BASE_BRANCH = $(DOCS_BASE_BRANCH)"

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
	@GO111MODULE=on go run -ldflags '-X main.cliName=ccloud' cmd/docs/main.go
	@GO111MODULE=on go run -ldflags '-X main.cliName=confluent' cmd/docs/main.go

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
	git commit -m "chore: update $(CLI_NAME) CLI docs for $(VERSION)" || exit 1; \
	git push origin $(CLI_NAME)-cli-$(VERSION) || exit 1; \
	hub pull-request -b $(DOCS_BASE_BRANCH) -m "chore: update $(CLI_NAME) CLI docs for $(VERSION)" || exit 1

.PHONY: clean-docs
clean-docs:
	@rm -rf docs/
