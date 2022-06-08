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
	git checkout -b update-whitelist-$(BUMPED_VERSION)
	
	go run -ldflags "-X main.version=$(BUMPED_VERSION)" cmd/usage/main.go > $(CC_CLI_SERVICE)/db/migrations/$(BUMPED_VERSION).up.sql
	echo "$$MIGRATEDOWN" > $(CC_CLI_SERVICE)/db/migrations/$(BUMPED_VERSION).down.sql
	
	cd $(CC_CLI_SERVICE) && \
	git add . && \
	git commit -m "update whitelist for $(BUMPED_VERSION)" && \
	git push origin update-whitelist-$(BUMPED_VERSION) && \
	hub pull-request -b master -m "update whitelist for $(BUMPED_VERSION)"
