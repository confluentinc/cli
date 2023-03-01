STAGING_BRANCH_REGEX="s/\([0-9]*\.[0-9]*\.\)[0-9]*/\1x/"
STAGING_BRANCH=$(shell echo $(CLEAN_VERSION) | sed $(STAGING_BRANCH_REGEX))

ifeq ($(DRY_RUN),true)
GIT_DRY_RUN_ARGS=--dry-run
endif

# The sed -i option operates differently on Mac vs Linux
OS:=$(shell uname)
ifeq ($(OS),Darwin)
SED_OPTION_INPLACE=-i ''
else
SED_OPTION_INPLACE=-i
endif

NEXT_MINOR_VERSION = $(word 1,$(split_version)).$(shell expr $(word 2,$(split_version)) + 1).0
SHORT_NEXT_MINOR_VERSION = $(word 1,$(split_version)).$(shell expr $(word 2,$(split_version)) + 1)
CURRENT_SHORT_MINOR_VERSION = $(word 1,$(split_version)).$(word 2,$(split_version))
NEXT_PATCH_VERSION = $(word 1,$(split_version)).$(word 2,$(split_version)).$(shell expr $(word 3,$(split_version)) + 1)

.PHONY: clone-docs-repos
clone-docs-repos:
	$(eval TMP_BASE=$(shell mktemp -d))
	$(eval CONFLUENT_DOCS_DIR=$(TMP_BASE)/docs-confluent-cli)
	git clone git@github.com:confluentinc/docs-confluent-cli.git $(CONFLUENT_DOCS_DIR)
	bump=$$(cat release-notes/bump.txt) && \
	version=$$(cat release-notes/version.txt) && \
	cd $(CONFLUENT_DOCS_DIR) && \
	git fetch && \
	if [ "$${bump}" = "patch" ]; then \
		git checkout $$(echo $${version} | sed $(STAGING_BRANCH_REGEX)) ; \
	fi

.PHONY: docs
docs: clean-docs
	go run cmd/docs/main.go

.PHONY: publish-docs
publish-docs: docs clone-docs-repos
	bump=$$(cat release-notes/bump.txt); \
	cd $(CONFLUENT_DOCS_DIR); \
	git checkout -b cli-$(VERSION) || exit 1; \
	rm -rf command-reference; \
	cp -R ~/git/go/src/github.com/confluentinc/cli/docs command-reference; \
	[ ! -f "command-reference/kafka/topic/confluent_kafka_topic_consume.rst" ] || sed -i '' 's/default "confluent_cli_consumer_[^"]*"/default "confluent_cli_consumer_<randomly-generated-id>"/' command-reference/kafka/topic/confluent_kafka_topic_consume.rst || exit 1; \
	git add . || exit 1; \
	git diff --cached --exit-code > /dev/null && echo "nothing to update for docs" && exit 0; \
	git commit -m "[ci skip] chore: update CLI docs for $(VERSION)" || exit 1; \
	git push $(GIT_DRY_RUN_ARGS) origin cli-$(VERSION) || exit 1; \
	base="master" && \
	if [ "$${bump}" = "patch" ]; then \
		base=$(STAGING_BRANCH) ; \
	fi && \
	gh pr create -B $${base} --title "chore: update CLI docs for $(VERSION)" --body "" || exit 1

.PHONY: clean-docs
clean-docs:
	rm -rf docs/

# This should be getting run after a version release has happened.
# So $(VERSION) is the version that was just released, and $$(cat release-notes/bump.txt)
# would be the next minor release (something in the future that doesn't exist yet).
.PHONY: release-docs
release-docs: clone-docs-repos cut-docs-branches update-settings-and-conf

# NB: When releasing a new version, the -post branch is updated to the current state of the .x branch,
# whether the -post branch exists or not. The `git checkout -B ...` handles this behavior.
.PHONY: cut-docs-branches
cut-docs-branches:
	cd $(CONFLUENT_DOCS_DIR) && \
	git fetch && \
	if [[ "$$(cat release-notes/bump.txt)" != "patch" ]]; then \
		git checkout master && \
		git checkout -b $(STAGING_BRANCH) && \
		git push $(GIT_DRY_RUN_ARGS) -u origin $(STAGING_BRANCH); \
	fi && \
	git checkout -B $(CURRENT_SHORT_MINOR_VERSION)-post origin/$(STAGING_BRANCH) && \
	git push $(GIT_DRY_RUN_ARGS) -u origin $(CURRENT_SHORT_MINOR_VERSION)-post

# If this is a patch release, we don't have to update master.
# The .x branch has been updated to ignore [ci skip] for minor branches since it is a pure upstream branch of minor branches.
# To ensure pint merge will work correctly, we will manually merge the -post branch into the .x branch using `-s ours`. Then, these
# branches will be pushed at the same time. This ensure there are no errors with pint merge.
.PHONY: update-settings-and-conf
update-settings-and-conf:
	cd $(CONFLUENT_DOCS_DIR) && \
	git fetch && \
	if [[ "$$(cat release-notes/bump.txt)" != "patch" ]]; then \
		git checkout master && \
		sed $(SED_OPTION_INPLACE) 's/export RELEASE_VERSION=.*/export RELEASE_VERSION=$(NEXT_MINOR_VERSION)-SNAPSHOT/g' settings.sh && \
		sed $(SED_OPTION_INPLACE) 's/export PUBLIC_VERSION=.*/export PUBLIC_VERSION=$(SHORT_NEXT_MINOR_VERSION)/g' settings.sh && \
		sed $(SED_OPTION_INPLACE) "s/^version = '.*'/version = \'$(SHORT_NEXT_MINOR_VERSION)\'/g" conf.py && \
		sed $(SED_OPTION_INPLACE) "s/^release = '.*'/release = \'$(NEXT_MINOR_VERSION)-SNAPSHOT\'/g" conf.py && \
		git commit -am "[ci skip] chore: update settings.sh and conf.py due to $(CLEAN_VERSION) release" && \
		git push $(GIT_DRY_RUN_ARGS); \
	fi && \
	git checkout $(STAGING_BRANCH) && \
	sed $(SED_OPTION_INPLACE) 's/export RELEASE_VERSION=.*/export RELEASE_VERSION=$(NEXT_PATCH_VERSION)-SNAPSHOT/g' settings.sh && \
	sed $(SED_OPTION_INPLACE) 's/export PUBLIC_VERSION=.*/export PUBLIC_VERSION=$(CURRENT_SHORT_MINOR_VERSION)/g' settings.sh && \
	sed $(SED_OPTION_INPLACE) "s/^version = '.*'/version = \'$(CURRENT_SHORT_MINOR_VERSION)\'/g" conf.py && \
	sed $(SED_OPTION_INPLACE) "s/^release = '.*'/release = \'$(NEXT_PATCH_VERSION)-SNAPSHOT\'/g" conf.py && \
	git commit -am "[ci skip] chore: update settings.sh and conf.py due to $(CLEAN_VERSION) release" && \
	git checkout $(CURRENT_SHORT_MINOR_VERSION)-post && \
	sed $(SED_OPTION_INPLACE) 's/export RELEASE_VERSION=.*/export RELEASE_VERSION=$(CLEAN_VERSION)/g' settings.sh && \
	sed $(SED_OPTION_INPLACE) 's/export PUBLIC_VERSION=.*/export PUBLIC_VERSION=$(CURRENT_SHORT_MINOR_VERSION)/g' settings.sh && \
	sed $(SED_OPTION_INPLACE) "s/^version = '.*'/version = \'$(CURRENT_SHORT_MINOR_VERSION)\'/g" conf.py && \
	sed $(SED_OPTION_INPLACE) "s/^release = '.*'/release = \'$(CLEAN_VERSION)\'/g" conf.py && \
	git commit -am "[ci skip] chore: update settings.sh and conf.py due to $(CLEAN_VERSION) release" && \
	git checkout $(STAGING_BRANCH) && \
	git merge -s ours -m "Merge branch '$(CURRENT_SHORT_MINOR_VERSION)-post' into $(STAGING_BRANCH)" $(CURRENT_SHORT_MINOR_VERSION)-post && \
	git push origin $(GIT_DRY_RUN_ARGS) "$(CURRENT_SHORT_MINOR_VERSION)-post:$(CURRENT_SHORT_MINOR_VERSION)-post" "$(STAGING_BRANCH):$(STAGING_BRANCH)"
