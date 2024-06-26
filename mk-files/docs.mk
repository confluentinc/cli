STAGING_BRANCH_REGEX="s/\([0-9]*\.[0-9]*\.\)[0-9]*/\1x/"
STAGING_BRANCH_REGEX_DASH="s/\([0-9]*\)\.\([0-9]*\)\.[0-9]*/\1-\2-x/"
STAGING_BRANCH=$(shell echo $(CLEAN_VERSION) | sed $(STAGING_BRANCH_REGEX))
STAGING_BRANCH_DASH=$(shell echo $(CLEAN_VERSION) | sed $(STAGING_BRANCH_REGEX_DASH))

ifeq ($(shell uname),Darwin)
SED ?= gsed
else
SED ?= sed
endif

NEXT_MINOR_VERSION = $(word 1,$(split_version)).$(shell expr $(word 2,$(split_version)) + 1).0
SHORT_NEXT_MINOR_VERSION = $(word 1,$(split_version)).$(shell expr $(word 2,$(split_version)) + 1)
CURRENT_SHORT_MINOR_VERSION = $(word 1,$(split_version)).$(word 2,$(split_version))
CURRENT_SHORT_MINOR_VERSION_DASH = $(word 1,$(split_version))-$(word 2,$(split_version))
NEXT_PATCH_VERSION = $(word 1,$(split_version)).$(word 2,$(split_version)).$(shell expr $(word 3,$(split_version)) + 1)

.PHONY: docs
docs:
	rm -rf docs/ && \
	go run cmd/docs/main.go

.PHONY: publish-docs
publish-docs:
	$(eval DIR=$(shell mktemp -d))
	$(eval CLI=$(DIR)/cli)
	$(eval CLI_RELEASE=$(DIR)/cli-release)
	$(eval DOCS_CONFLUENT_CLI=$(DIR)/docs-confluent-cli)

	git clone git@github.com:confluentinc/cli-release.git $(CLI_RELEASE) && \
	go run $(CLI_RELEASE)/cmd/releasenotes/formatter/main.go $(CLI_RELEASE)/release-notes/$(CLEAN_VERSION).json docs > $(DIR)/release-notes.rst && \
	git clone git@github.com:confluentinc/cli.git $(CLI) && \
	cd $(CLI) && \
	go run cmd/docs/main.go && \
	cd - && \
	git clone git@github.com:confluentinc/docs-confluent-cli.git $(DOCS_CONFLUENT_CLI) && \
	cd $(DOCS_CONFLUENT_CLI) && \
	git checkout -b publish-docs-v$(CLEAN_VERSION) && \
	$(SED) -i "10r $(DIR)/release-notes.rst" release-notes.rst && \
	rm -rf command-reference && \
	cp -R $(CLI)/docs command-reference && \
	git add . && \
	git commit --allow-empty -m "[pint skip] chore: update CLI docs for v$(CLEAN_VERSION)" && \
	$(call dry-run,git push origin publish-docs-v$(CLEAN_VERSION)) && \
	base="master" && \
	$(call dry-run,gh pr create --base $${base} --title "chore: update CLI docs for v$(CLEAN_VERSION)" --body "")

	rm -rf $(DIR)

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
		git checkout -b $(CURRENT_SHORT_MINOR_VERSION)-post origin/$(STAGING_BRANCH) && \
		$(call dry-run,git push -u origin $(CURRENT_SHORT_MINOR_VERSION)-post) && \
		git checkout master && \
		git checkout -b release-docs-$(CLEAN_VERSION) && \
		$(SED) -i 's/export RELEASE_VERSION=.*/export RELEASE_VERSION=$(NEXT_MINOR_VERSION)-SNAPSHOT/g' settings.sh && \
		$(SED) -i 's/export PUBLIC_VERSION=.*/export PUBLIC_VERSION=$(SHORT_NEXT_MINOR_VERSION)/g' settings.sh && \
		$(SED) -i "s/^version = '.*'/version = \'$(SHORT_NEXT_MINOR_VERSION)\'/g" conf.py && \
		$(SED) -i "s/^release = '.*'/release = \'$(NEXT_MINOR_VERSION)-SNAPSHOT\'/g" conf.py && \
		git commit -am "[pint skip] chore: update settings.sh and conf.py due to $(CLEAN_VERSION) release" && \
		$(call dry-run,git push origin release-docs-$(CLEAN_VERSION)) && \
		$(call dry-run,gh pr create --base master --title "Release docs for v$(CLEAN_VERSION)" --body ""); \
	else \
		git checkout -b release-docs-$(STAGING_BRANCH_DASH) origin/$(STAGING_BRANCH) && \
		git checkout origin/master -- command-reference release-notes.rst && \
		git commit -am "Pull updated command-reference and release-notes.rst from master" && \
		$(call dry-run,git push origin release-docs-$(STAGING_BRANCH_DASH)) && \
		$(call dry-run,gh pr create --base $(STAGING_BRANCH) --title "Release docs for v$(CLEAN_VERSION)" --body ""); \
		git checkout -b release-docs-$(CURRENT_SHORT_MINOR_VERSION_DASH) origin/$(CURRENT_SHORT_MINOR_VERSION)-post && \
		git checkout origin/master -- command-reference release-notes.rst && \
		git commit -am "Pull updated command-reference and release-notes.rst from master" && \
		$(call dry-run,git push origin release-docs-$(CURRENT_SHORT_MINOR_VERSION_DASH)) && \
		$(call dry-run,gh pr create --base $(CURRENT_SHORT_MINOR_VERSION)-post --title "Release docs for v$(CLEAN_VERSION)" --body ""); \
	fi

	rm -rf $(DIR)

.PHONY: release-docs-staging
release-docs-staging:
	$(eval DIR=$(shell mktemp -d))
	$(eval DOCS_CONFLUENT_CLI=$(DIR)/docs-confluent-cli)

	git clone git@github.com:confluentinc/docs-confluent-cli.git $(DOCS_CONFLUENT_CLI) && \
	cd $(DOCS_CONFLUENT_CLI) && \
	git fetch && \
	git checkout -b release-docs-staging-$(CURRENT_SHORT_MINOR_VERSION_DASH) origin/$(CURRENT_SHORT_MINOR_VERSION)-post && \
	$(SED) -i 's/export RELEASE_VERSION=.*/export RELEASE_VERSION=$(CLEAN_VERSION)/g' settings.sh && \
	$(SED) -i 's/export PUBLIC_VERSION=.*/export PUBLIC_VERSION=$(CURRENT_SHORT_MINOR_VERSION)/g' settings.sh && \
	$(SED) -i "s/^version = '.*'/version = \'$(CURRENT_SHORT_MINOR_VERSION)\'/g" conf.py && \
	$(SED) -i "s/^release = '.*'/release = \'$(CLEAN_VERSION)\'/g" conf.py && \
	git commit -am "[pint skip] chore: update settings.sh and conf.py due to $(CLEAN_VERSION) release" && \
	$(call dry-run,git push origin release-docs-staging-$(CURRENT_SHORT_MINOR_VERSION_DASH)) && \
	$(call dry-run,gh pr create --base $(CURRENT_SHORT_MINOR_VERSION)-post --title "[pint skip] Release docs to staging for v$(CLEAN_VERSION) release" --body "") && \
	git checkout -b release-docs-staging-$(STAGING_BRANCH_DASH) origin/$(STAGING_BRANCH) && \
	$(SED) -i 's/export RELEASE_VERSION=.*/export RELEASE_VERSION=$(NEXT_PATCH_VERSION)-SNAPSHOT/g' settings.sh && \
	$(SED) -i 's/export PUBLIC_VERSION=.*/export PUBLIC_VERSION=$(CURRENT_SHORT_MINOR_VERSION)/g' settings.sh && \
	$(SED) -i "s/^version = '.*'/version = \'$(CURRENT_SHORT_MINOR_VERSION)\'/g" conf.py && \
	$(SED) -i "s/^release = '.*'/release = \'$(NEXT_PATCH_VERSION)-SNAPSHOT\'/g" conf.py && \
	git commit -am "[pint skip] chore: update settings.sh and conf.py due to $(CLEAN_VERSION) release" && \
	git merge -s ours -m "Merge branch 'release-docs-staging-$(CURRENT_SHORT_MINOR_VERSION_DASH)' into release-docs-staging-$(STAGING_BRANCH_DASH)" release-docs-staging-$(CURRENT_SHORT_MINOR_VERSION_DASH) && \
	$(call dry-run,git push origin release-docs-staging-$(STAGING_BRANCH_DASH)) && \
	$(call dry-run,gh pr create --base $(STAGING_BRANCH) --title "[pint skip] Release docs to staging for v$(CLEAN_VERSION) release" --body "")

	rm -rf $(DIR)
