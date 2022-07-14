define MIGRATEDOWN
BEGIN;

DELETE FROM whitelist WHERE version = '$(BUMPED_VERSION)';

COMMIT;
endef

export MIGRATEDOWN
update-whitelist:
	$(eval CC_CLI_SERVICE=$(shell mktemp -d)/cc-cli-service)
	
	git clone git@github.com:confluentinc/cc-cli-service.git $(CC_CLI_SERVICE) && \
	cd $(CC_CLI_SERVICE) && \
	git checkout -b update-whitelist-$(BUMPED_VERSION) && \
	make db-migrate-create NAME=$(BUMPED_VERSION)
	
	go run -ldflags "-X main.version=$(BUMPED_VERSION)" cmd/usage/main.go > $$(find $(CC_CLI_SERVICE)/db/migrations/ -name "*_$(BUMPED_VERSION).up.sql")
	echo "$$MIGRATEDOWN" > $$(find $(CC_CLI_SERVICE)/db/migrations/ -name "*_$(BUMPED_VERSION).down.sql")
	
	cd $(CC_CLI_SERVICE) && \
	make db-migrate-up && \
	git add . && \
	git commit -m "update whitelist for $(BUMPED_VERSION)" && \
	git push origin update-whitelist-$(BUMPED_VERSION) && \
	gh pr create -B master -t "Update whitelist for $(BUMPED_VERSION)"
