# If you set up your laptop following https://github.com/confluentinc/cc-documentation/blob/master/Operations/Laptop%20Setup.md
# then assuming caas.sh lives here should be fine
define aws-authenticate
	$(call dry-run,export AWS_PROFILE=cc-production-1/prod-administrator)
endef

.PHONY: update-package-managers
update-package-managers:
	VERSION=$(VERSION) scripts/build_linux.sh && \
	$(call dry-run, aws s3 sync deb $(S3_DEB_RPM_PROD_PATH)/deb) && \
	$(call dry-run, aws s3 sync rpm $(S3_DEB_RPM_PROD_PATH)/rpm) && \
	$(call dry-run, s3-repo-utils -v website index --fake-index --prefix $(S3_DEB_RPM_PROD_PREFIX)/ $(S3_DEB_RPM_BUCKET_NAME))

.PHONY: update-muckrake
update-muckrake:
	$(eval DIR=$(shell mktemp -d))
	$(eval CLI_RELEASE=$(DIR)/cli-release)
	$(eval MUCKRAKE=$(DIR)/muckrake)

	git clone git@github.com:confluentinc/cli-release.git $(CLI_RELEASE) && \
	version=$$(ls $(CLI_RELEASE)/release-notes | $(SED) "s/.json$$//" | sort --version-sort | tail -1) && \
	git clone git@github.com:confluentinc/muckrake.git $(MUCKRAKE) && \
	cd $(MUCKRAKE) && \
	base=$$(git branch --remote --format "%(refname:short)" | sed -n "s|^origin/\([1-9][0-9]*\.[0-9][0-9]*\.x\)$$|\1|p" | tail -1) && \
	git checkout $$base && \
	branch=bump-cli && \
	git checkout -b $$branch && \
	$(SED) -i "s|confluent-cli-.*=\$${confluent_s3}/confluent\.cloud/confluent-cli/archives/.*/confluent_.*_linux_amd64\.tar\.gz|confluent-cli-$${version}=\$${confluent_s3}/confluent.cloud/confluent-cli/archives/$${version}/confluent_$${version}_linux_amd64.tar.gz|" ducker/ducker && \
	$(SED) -i "s|VERSION = \".*\"|VERSION = \"$${version}\"|" muckrake/services/cli.py && \
	$(SED) -i "s|get_cli .*|get_cli $${version}|" vagrant/base-redhat.sh && \
	$(SED) -i "s|get_cli .*|get_cli $${version}|" vagrant/base-redhat8.sh && \
	$(SED) -i "s|get_cli .*|get_cli $${version}|" vagrant/base-redhat9.sh && \
	$(SED) -i "s|get_cli .*|get_cli $${version}|" vagrant/base-ubuntu.sh && \
	git commit -am "bump cli to v$${version}" || true && \
	$(call dry-run,git push --force --set-upstream origin $$branch) && \
	if gh pr view $$branch --json state --jq .state 2>&1 | grep -E "no pull requests found|CLOSED|MERGED"; then \
		$(call dry-run,gh pr create --base $${base} --title "Bump CLI to v$${version}" --body "") && \
		$(call dry-run,gh pr merge --squash --auto); \
	fi

	rm -rf $(DIR)

.PHONY: update-packaging
update-packaging:
	$(eval DIR=$(shell mktemp -d))
	$(eval CLI_RELEASE=$(DIR)/cli-release)
	$(eval PACKAGING=$(DIR)/packaging)

	git clone git@github.com:confluentinc/cli-release.git $(CLI_RELEASE) && \
	version=$$(ls $(CLI_RELEASE)/release-notes | $(SED) "s/.json$$//" | sort --version-sort | tail -1) && \
	git clone git@github.com:confluentinc/packaging.git $(PACKAGING) && \
	cd $(PACKAGING) && \
	base=$$(git branch --remote --format "%(refname:short)" | sed -n "s|^origin/\([1-9][0-9]*\.[0-9][0-9]*\.x\)$$|\1|p" | tail -1) && \
	git checkout $$base && \
	branch="bump-cli" && \
	git checkout -b $$branch && \
	$(SED) -i "s|cli_BRANCH=\".*\"|cli_BRANCH=\"v$${version}\"|" settings.sh && \
	$(SED) -i "s|CLI_VERSION=.*|CLI_VERSION=$${version}|" release_testing/bin/smoke_test.sh && \
	$(SED) -i "s|cli: \".*\"|cli: \"v$${version}\"|" packaging/resources/packaging-run-config.yml && \
	git commit -am "bump cli to v$${version}" && \
	$(call dry-run,git push --force --set-upstream origin $$branch) && \
	if gh pr view $$branch --json state --jq .state 2>&1 | grep -E "no pull requests found|CLOSED|MERGED"; then \
		$(call dry-run,gh pr create --base $${base} --title "Bump CLI to v$${version}" --body "") && \
		$(call dry-run,gh pr merge --squash --auto); \
	fi

	rm -rf $(DIR)
