.PHONY: cpd-debug-and-err
cpd-debug-and-err:
	$(CPD_PATH) debug --id `kubectl config current-context`; exit 1

.PHONY: cpd-priv-testenv
cpd-priv-testenv:
	@echo "## Exporting CPD environment bash profile."
	set -o pipefail && $(CPD_PATH) priv testenv --id `kubectl config current-context` > $(CC_SYSTEM_TEST_ENV_SECRETS)
	
.PHONY: system-test-init-env
system-test-init-env:
	source $(CC_SYSTEM_TEST_ENV_SECRETS) && $(MAKE) -C $(CC_SYSTEM_TEST_CHECKOUT_DIR) init-env

.PHONY: run-system-tests
run-system-tests:
	source $(CC_SYSTEM_TEST_ENV_SECRETS) && TEST_REPORT_FILE="$(BUILD_DIR)/ci-gating/TEST-report.xml" $(MAKE) -C $(CC_SYSTEM_TEST_CHECKOUT_DIR) test

.PHONY: replace-cli-binary
replace-cli-binary:
	echo $$(ls)
	cp ./dist/ccloud/ccloud_linux_amd64/ccloud $(CC_SYSTEM_TEST_CHECKOUT_DIR)/test/cli/cli_bin/linux_amd64/ccloud 
