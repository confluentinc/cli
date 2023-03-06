update-whitelist:
	$(eval CC_CLI_SERVICE=$(shell mktemp -d)/cc-cli-service)
	
	bump=$$(cat release-notes/bump.txt) && \
	version=$$(cat release-notes/version.txt) && \
	git clone git@github.com:confluentinc/cc-cli-service.git $(CC_CLI_SERVICE) && \
	cd $(CC_CLI_SERVICE) && \
	make db-local-reset && \
	git checkout -b cli-v$${version} && \
	make db-migrate-create NAME=v$${version} && \
	cd - && \
	echo -e "BEGIN;\n\nDELETE FROM whilelist WHERE version = 'v$${version}';\n\nCOMMIT;\n" > $$(find $(CC_CLI_SERVICE)/db/migrations/ -name "*_v$${version}.down.sql") && \
	go run -ldflags "-X main.version=v$${version}" cmd/usage/main.go > $$(find $(CC_CLI_SERVICE)/db/migrations/ -name "*_v$${version}.up.sql") && \
	cd $(CC_CLI_SERVICE) && \
	git add . && \
	git commit -m "[ci skip] update whitelist for v$${version}" && \
	cd db/migrations/ && \
	a=$$(ls | grep up | tail -n 2 | head -n 1) && \
	b=$$(ls | grep up | tail -n 1) && \
	sed -i "" "s/v[0-9]*\.[0-9]*\.[0-9]*/v$${version}/" $$a && \
	body=$$(echo -e "\`\`\`diff\n$$(diff -u $$a $$b)\n\`\`\`") && \
	git push origin cli-v$${version} && \
	gh pr create -B master --title "[ci skip] Update whitelist for v$${version}" --body "$${body}"

update-db:
	$(eval CC_CLI_SERVICE=$(shell mktemp -d)/cc-cli-service)

	bump=$$(cat release-notes/bump.txt) && \
	version=$$(cat release-notes/version.txt) && \
	source ~/git/go/src/github.com/confluentinc/cc-dotfiles/caas.sh && \
	git clone git@github.com:confluentinc/cc-cli-service.git $(CC_CLI_SERVICE) && \
	cd $(CC_CLI_SERVICE) && \
	git checkout -b cli-v$${version} && \
	sed -i "" "s|db.url: .*|db.url: postgres://cc_cli_service@127.0.0.1:8432/cli?sslmode=require|" config.yaml && \
	for pair in stag,237597620434 devel,037803949979 prod,050879227952; do \
		IFS="," read env arn <<< "$$pair"; \
		eval $$(gimme-aws-creds --output-format export --roles arn:aws:iam::$${arn}:role/administrator); \
		cctunnel -e $$env -b cli -i read-write; \
		sleep 3; \
		make db-migrate-up; \
		kill -9 $$(lsof -i 4:8432 | awk 'NR > 1 { print $$2 };'); \
	done && \
	git add db/schema.sql && \
	git commit -m "[ci skip] update db for v$${version}" && \
	git push origin cli-v$${version} && \
	gh pr create -B master --title "[ci skip] Update DB for v$${version}" --body ""

promote:
	$(eval DIR=$(shell mktemp -d))
	$(eval CC_CLI_SERVICE=$(DIR)/cc-cli-service)
	
	git clone git@github.com:confluentinc/cc-cli-service.git $(CC_CLI_SERVICE) && \
	cd $(CC_CLI_SERVICE) && \
	git checkout -b promote && \
	export VAULT_ADDR=https://vault.cireops.gcp.internal.confluent.cloud; vault login -method=oidc -path=okta/ && \
	halctl --context prod --vault-oidc-role halyard-prod --vault-token "$$(cat ~/.vault-token)" --vault-login-path "auth/app/prod/login" release service environment version list cc-cli-service stag | sed -n 's/\*.*\(0.[0-9]*.0\).*/\1/p' > $(DIR)/version.txt && \
	for env in devel prod; do \
		halctl --context prod --vault-oidc-role halyard-prod --vault-token "$$(cat ~/.vault-token)" --vault-login-path "auth/app/prod/login" release service environment version list cc-cli-service $${env} | grep $$(cat $(DIR)/version.txt) | grep -o -E '[0-9]+' | head -1 > $(DIR)/$${env}.txt; \
		sed -i '' "s/installedVersion: \"[0-9]*\"/installedVersion: \"$$(cat $(DIR)/$${env}.txt)\"/" .deployed-versions/$${env}.yaml; \
	done && \
	git commit -am "promote devel to $$(cat $(DIR)/devel.txt) and promote prod to $$(cat $(DIR)/prod.txt)" && \
	git push origin promote && \
	gh pr create -B master --title "Promote devel and prod" --body ""
