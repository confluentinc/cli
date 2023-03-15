update-db:
	$(eval DIR=$(shell mktemp -d))
	$(eval CC_CLI_SERVICE=$(DIR)/cc-cli-service)
	
	version=$$(cat release-notes/version.txt) && \
	git clone git@github.com:confluentinc/cc-cli-service.git $(CC_CLI_SERVICE) && \
	cd $(CC_CLI_SERVICE) && \
	make db-local-reset && \
	git checkout -b update-db-v$${version} && \
	make db-migrate-create NAME=v$${version} && \
	cd - && \
	echo -e "BEGIN;\n\nDELETE FROM whilelist WHERE version = 'v$${version}';\n\nCOMMIT;\n" > $$(find $(CC_CLI_SERVICE)/db/migrations/ -name "*_v$${version}.down.sql") && \
	go run -ldflags "-X main.version=v$${version}" cmd/usage/main.go > $$(find $(CC_CLI_SERVICE)/db/migrations/ -name "*_v$${version}.up.sql") && \
	cd $(CC_CLI_SERVICE) && \
	make db-migrate-up && \
	git add . && \
	git commit -m "update db for v$${version}" && \
	cd db/migrations/ && \
	a=$$(ls | grep up | tail -n 2 | head -n 1) && \
	b=$$(ls | grep up | tail -n 1) && \
	sed -i "" "s/v[0-9]*\.[0-9]*\.[0-9]*/v$${version}/" $$a && \
	diff=$$(diff -u $$a $$b) && \
	body=$$(echo -e "\`\`\`diff$${diff}\n\n\`\`\`") && \
	if [ "$${diff}" = "" ]; then \
		body="No changes."; \
	fi && \
	$(call dry-run,git push origin update-db-v$${version}) && \
	$(call dry-run,gh pr create -B master --title "Update DB for v$${version}" --body "$${body}")

	rm -rf $(DIR)

promote:
	$(eval DIR=$(shell mktemp -d))
	$(eval CC_CLI_SERVICE=$(DIR)/cc-cli-service)
	
	git clone git@github.com:confluentinc/cc-cli-service.git $(CC_CLI_SERVICE) && \
	cd $(CC_CLI_SERVICE) && \
	export VAULT_ADDR=https://vault.cireops.gcp.internal.confluent.cloud && \
	vault login -method=oidc -path=okta && \
	halctl --context prod --vault-oidc-role halyard-prod --vault-token "$$(cat ~/.vault-token)" --vault-login-path "auth/app/prod/login" release service environment version list cc-cli-service stag | sed -n 's/\*.*\(0.[0-9]*.0\).*/\1/p' > $(DIR)/stag.txt && \
	git checkout -b promote-$$(cat $(DIR)/stag.txt) && \
	for env in devel prod; do \
		halctl --context prod --vault-oidc-role halyard-prod --vault-token "$$(cat ~/.vault-token)" --vault-login-path "auth/app/prod/login" release service environment version list cc-cli-service $${env} | grep $$(cat $(DIR)/stag.txt) | grep -o -E '[0-9]+' | head -1 > $(DIR)/$${env}.txt; \
		sed -i '' "s/installedVersion: \"[0-9]*\"/installedVersion: \"$$(cat $(DIR)/$${env}.txt)\"/" .deployed-versions/$${env}.yaml; \
	done && \
	git commit -am "promote devel and prod" && \
	$(call dry-run,git push origin promote-$$(cat $(DIR)/stag.txt)) && \
	$(call dry-run,gh pr create -B master --title "Promote devel and prod" --body "")

	rm -rf $(DIR)
