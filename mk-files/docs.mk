STAGING_BRANCH_REGEX="s/\([0-9]*\.[0-9]*\.\)[0-9]*/\1x/"
STAGING_BRANCH=$(shell echo $(CLEAN_VERSION) | sed $(STAGING_BRANCH_REGEX))

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

.PHONY: docs
docs:
	rm -rf docs/ && \
	go run cmd/docs/main.go

.PHONY: publish-docs
publish-docs: docs
	$(eval DIR=$(shell mktemp -d))
	$(eval DOCS_CONFLUENT_CLI=$(DIR)/docs-confluent-cli)

	git clone git@github.com:confluentinc/docs-confluent-cli.git $(DOCS_CONFLUENT_CLI) && \
	version=$$(cat release-notes/version.txt) && \
	cd $(DOCS_CONFLUENT_CLI) && \
	git checkout publish-docs-$${version} && \
	rm -rf command-reference && \
	cp -R ~/git/go/src/github.com/confluentinc/cli/docs command-reference && \
	[ ! -f "command-reference/kafka/topic/confluent_kafka_topic_consume.rst" ] || sed -i '' 's/default "confluent_cli_consumer_[^"]*"/default "confluent_cli_consumer_<randomly-generated-id>"/' command-reference/kafka/topic/confluent_kafka_topic_consume.rst || exit 1 && \
	git add . && \
	git diff --cached --exit-code > /dev/null && echo "nothing to update for docs" && exit 0; \
	git commit -m "[ci skip] chore: update CLI docs for $${version}" && \
	$(call dry-run,git push origin publish-docs-$${version}) && \
	base="master" && \
	if [[ $${version} != *.0 ]]; then \
		base=$(STAGING_BRANCH); \
	fi && \
	$(call dry-run,gh pr create -B $${base} --title "chore: update CLI docs for $${version}" --body "")

	rm -rf $(DIR)

# NB: When releasing a new version, the -post branch is updated to the current state of the .x branch,
# whether the -post branch exists or not. The `git checkout -B ...` handles this behavior.
# If this is a patch release, we don't have to update master.
# The .x branch has been updated to ignore [ci skip] for minor branches since it is a pure upstream branch of minor branches.
# To ensure pint merge will work correctly, we will manually merge the -post branch into the .x branch using `-s ours`. Then, these
# branches will be pushed at the same time. This ensure there are no errors with pint merge.
.PHONY: release-docs
release-docs:
	$(eval DIR=$(shell mktemp -d))
	$(eval DOCS_CONFLUENT_CLI=$(DIR)/docs-confluent-cli)

	git clone git@github.com:confluentinc/docs-confluent-cli.git $(DOCS_CONFLUENT_CLI) && \
	version=$$(cat release-notes/version.txt) && \
	cd $(DOCS_CONFLUENT_CLI) && \
	git fetch && \
	if [[ $${version} == *.0 ]]; then \
		git checkout master && \
		git checkout -b $(STAGING_BRANCH) && \
		$(call dry-run,git push -u origin $(STAGING_BRANCH)); \
	fi && \
	git checkout -B $(CURRENT_SHORT_MINOR_VERSION)-post origin/$(STAGING_BRANCH) && \
	$(call dry-run,git push -u origin $(CURRENT_SHORT_MINOR_VERSION)-post) && \
	if [[ $${version} == *.0 ]]; then \
		git checkout master && \
		sed $(SED_OPTION_INPLACE) 's/export RELEASE_VERSION=.*/export RELEASE_VERSION=$(NEXT_MINOR_VERSION)-SNAPSHOT/g' settings.sh && \
		sed $(SED_OPTION_INPLACE) 's/export PUBLIC_VERSION=.*/export PUBLIC_VERSION=$(SHORT_NEXT_MINOR_VERSION)/g' settings.sh && \
		sed $(SED_OPTION_INPLACE) "s/^version = '.*'/version = \'$(SHORT_NEXT_MINOR_VERSION)\'/g" conf.py && \
		sed $(SED_OPTION_INPLACE) "s/^release = '.*'/release = \'$(NEXT_MINOR_VERSION)-SNAPSHOT\'/g" conf.py && \
		git commit -am "[ci skip] chore: update settings.sh and conf.py due to $(CLEAN_VERSION) release" && \
		$(call dry-run,git push); \
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
	$(call dry-run,git push origin "$(CURRENT_SHORT_MINOR_VERSION)-post:$(CURRENT_SHORT_MINOR_VERSION)-post" "$(STAGING_BRANCH):$(STAGING_BRANCH)")

	rm -rf $(DIR)
