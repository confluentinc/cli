promote:
	$(eval DIR=$(shell mktemp -d))
	$(eval CC_CLI_SERVICE=$(DIR)/cc-cli-service)
	
	git clone git@github.com:confluentinc/cc-cli-service.git $(CC_CLI_SERVICE) && \
	cd $(CC_CLI_SERVICE) && \
	export VAULT_ADDR=https://vault.cireops.gcp.internal.confluent.cloud && \
	vault login -method=oidc -path=okta && \
	halctl --context prod --vault-oidc-role halyard-prod --vault-token "$$(cat ~/.vault-token)" --vault-login-path "auth/app/prod/login" release service environment version list cc-cli-service stag | sed -n 's/\*.*\(0\.[0-9]*\.0\).*/\1/p' > $(DIR)/stag.txt && \
	git checkout -b promote-$$(cat $(DIR)/stag.txt) && \
	for env in devel prod; do \
		halctl --context prod --vault-oidc-role halyard-prod --vault-token "$$(cat ~/.vault-token)" --vault-login-path "auth/app/prod/login" release service environment version list cc-cli-service $${env} | grep $$(cat $(DIR)/stag.txt) | grep -o -E '[0-9]+' | head -1 > $(DIR)/$${env}.txt; \
		sed -i '' "s/installedVersion: \"[0-9]*\"/installedVersion: \"$$(cat $(DIR)/$${env}.txt)\"/" .deployed-versions/$${env}.yaml; \
	done && \
	git commit -am "promote devel and prod" && \
	$(call dry-run,git push origin promote-$$(cat $(DIR)/stag.txt)) && \
	$(call dry-run,gh pr create --base master --title "Promote devel and prod" --body "")

	rm -rf $(DIR)
