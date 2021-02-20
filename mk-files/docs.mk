DOCS_REPOS=docs-ccloud-cli docs-confluent-cli

DOCS_BASE_BRANCH=unset
ifeq ($(BUMP),auto)
DOCS_BASE_BRANCH=master
else ifeq ($(BUMP), major)
DOCS_BASE_BRANCH=master
else ifeq ($(BUMP), minor)
DOCS_BASE_BRANCH=master
else ifeq ($(BUMP), patch)
DOCS_BASE_BRANCH=$(shell echo $(BUMPED_CLEAN_VERSION) | sed 's/\([0-9]*\.[0-9]*\.\)[0-9]*/\1x/')
endif

.PHONY: cut-docs-branches
cut-docs-branches:
	echo "DOCS_BASE_BRANCH = $(DOCS_BASE_BRANCH)"

.PHONY: clone-docs-repos
	$(eval TMP_BASE=$(shell mktemp -d))
	$(eval TMP_DOCS_CCLOUD=$(TMP_BASE)/docs-ccloud-cli)
	$(eval TMP_DOCS_CONFLUENT=$(TMP_BASE)/docs-confluent-cli)
	git clone git@github.com:confluentinc/docs-ccloud-cli.git $(TMP_DOCS_CCLOUD)
	git clone git@github.com:confluentinc/docs-confluent-cli.git $(TMP_DOCS_CONFLUENT)
	for repo in $(TMP_DOCS_CCLOUD) $(TMP_DOCS_CONLUFNET) ; do \
		cd $${repo} && \
		git fetch && \
		git checkout $(DOCS_BASE_BRANCH) && \
		cd .. ; \
	done
