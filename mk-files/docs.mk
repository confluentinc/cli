STAGING_BRANCH_REGEX="s/\([0-9]*\.[0-9]*\.\)[0-9]*/\1x/"
STAGING_BRANCH=$(shell echo $(CLEAN_VERSION) | sed $(STAGING_BRANCH_REGEX))

ifeq ($(shell uname),Darwin)
SED ?= gsed
else
SED ?= sed
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
	$(eval CLI_RELEASE=$(DIR)/cli-release)
	$(eval DOCS_CONFLUENT_CLI=$(DIR)/docs-confluent-cli)

	git clone git@github.com:confluentinc/docs-confluent-cli.git $(DOCS_CONFLUENT_CLI) && \
	cd $(DOCS_CONFLUENT_CLI) && \
	git checkout -b publish-docs-v$(CLEAN_VERSION) && \
	cd - && \
	git clone git@github.com:confluentinc/cli-release.git $(CLI_RELEASE) && \
	go run $(CLI_RELEASE)/cmd/releasenotes/formatter/main.go $(CLI_RELEASE)/release-notes/$(CLEAN_VERSION).json docs > $(DIR)/release-notes.rst && \
	$(SED) -i "10r $(DIR)/release-notes.rst" $(DOCS_CONFLUENT_CLI)/release-notes.rst && \
	cd $(DOCS_CONFLUENT_CLI) && \
	rm -rf command-reference && \
	cp -R ~/git/go/src/github.com/confluentinc/cli/docs command-reference && \
	git add . && \
	git commit --allow-empty -m "[ci skip] chore: update CLI docs for v$(CLEAN_VERSION)" && \
	$(call dry-run,git push origin publish-docs-v$(CLEAN_VERSION)) && \
	base="master" && \
	if [[ $(CLEAN_VERSION) != *.0 ]]; then \
		base=$(STAGING_BRANCH); \
	fi && \
	$(call dry-run,gh pr create --base $${base} --title "chore: update CLI docs for v$(CLEAN_VERSION)" --body "")

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
	cd $(DOCS_CONFLUENT_CLI) && \
	git fetch && \
	if [[ $(CLEAN_VERSION) == *.0 ]]; then \
		git checkout master && \
		git checkout -b $(STAGING_BRANCH) && \
		$(call dry-run,git push -u origin $(STAGING_BRANCH)); \
	fi && \
	git checkout -B $(CURRENT_SHORT_MINOR_VERSION)-post origin/$(STAGING_BRANCH) && \
	$(call dry-run,git push -u origin $(CURRENT_SHORT_MINOR_VERSION)-post) && \
	if [[ $(CLEAN_VERSION) == *.0 ]]; then \
		git checkout master && \
		$(SED) -i 's/export RELEASE_VERSION=.*/export RELEASE_VERSION=$(NEXT_MINOR_VERSION)-SNAPSHOT/g' settings.sh && \
		$(SED) -i 's/export PUBLIC_VERSION=.*/export PUBLIC_VERSION=$(SHORT_NEXT_MINOR_VERSION)/g' settings.sh && \
		$(SED) -i "s/^version = '.*'/version = \'$(SHORT_NEXT_MINOR_VERSION)\'/g" conf.py && \
		$(SED) -i "s/^release = '.*'/release = \'$(NEXT_MINOR_VERSION)-SNAPSHOT\'/g" conf.py && \
		git commit -am "[ci skip] chore: update settings.sh and conf.py due to $(CLEAN_VERSION) release" && \
		$(call dry-run,git push); \
	fi && \
	git checkout $(STAGING_BRANCH) && \
	$(SED) -i 's/export RELEASE_VERSION=.*/export RELEASE_VERSION=$(NEXT_PATCH_VERSION)-SNAPSHOT/g' settings.sh && \
	$(SED) -i 's/export PUBLIC_VERSION=.*/export PUBLIC_VERSION=$(CURRENT_SHORT_MINOR_VERSION)/g' settings.sh && \
	$(SED) -i "s/^version = '.*'/version = \'$(CURRENT_SHORT_MINOR_VERSION)\'/g" conf.py && \
	$(SED) -i "s/^release = '.*'/release = \'$(NEXT_PATCH_VERSION)-SNAPSHOT\'/g" conf.py && \
	git commit -am "[ci skip] chore: update settings.sh and conf.py due to $(CLEAN_VERSION) release" && \
	git checkout $(CURRENT_SHORT_MINOR_VERSION)-post && \
	$(SED) -i 's/export RELEASE_VERSION=.*/export RELEASE_VERSION=$(CLEAN_VERSION)/g' settings.sh && \
	$(SED) -i 's/export PUBLIC_VERSION=.*/export PUBLIC_VERSION=$(CURRENT_SHORT_MINOR_VERSION)/g' settings.sh && \
	$(SED) -i "s/^version = '.*'/version = \'$(CURRENT_SHORT_MINOR_VERSION)\'/g" conf.py && \
	$(SED) -i "s/^release = '.*'/release = \'$(CLEAN_VERSION)\'/g" conf.py && \
	git commit -am "[ci skip] chore: update settings.sh and conf.py due to $(CLEAN_VERSION) release" && \
	git checkout $(STAGING_BRANCH) && \
	git merge -s ours -m "Merge branch '$(CURRENT_SHORT_MINOR_VERSION)-post' into $(STAGING_BRANCH)" $(CURRENT_SHORT_MINOR_VERSION)-post && \
	$(call dry-run,git push origin "$(CURRENT_SHORT_MINOR_VERSION)-post:$(CURRENT_SHORT_MINOR_VERSION)-post" "$(STAGING_BRANCH):$(STAGING_BRANCH)")

	rm -rf $(DIR)
