JFROG_DOCKER_REPO := confluent-docker.jfrog.io
JFROG_DOCKER_REPO_INTERNAL := confluent-docker-internal-dev.jfrog.io

.PHONY: docker-login
## Login to docker Artifactory
docker-login:
ifeq ($(DOCKER_USER)$(DOCKER_APIKEY),$(_empty))
	@jq -e '.auths."$(JFROG_DOCKER_REPO)"' $(HOME)/.docker/config.json 2>&1 >/dev/null ||\
		(echo "$(JFROG_DOCKER_REPO) not logged in, Username and Password not found in environment, prompting for login:" && \
		 docker login $(JFROG_DOCKER_REPO))
	@jq -e '.auths."$(JFROG_DOCKER_REPO_INTERNAL)"' $(HOME)/.docker/config.json 2>&1 >/dev/null ||\
		(echo "$(JFROG_DOCKER_REPO_INTERNAL) not logged in, Username and Password not found in environment, prompting for login:" && \
		 docker login $(JFROG_DOCKER_REPO_INTERNAL))
else
	@jq -e '.auths."$(JFROG_DOCKER_REPO)"' $(HOME)/.docker/config.json 2>&1 >/dev/null ||\
		docker login $(JFROG_DOCKER_REPO) --username $(DOCKER_USER) --password $(DOCKER_APIKEY)
	@jq -e '.auths."$(JFROG_DOCKER_REPO_INTERNAL)"' $(HOME)/.docker/config.json 2>&1 >/dev/null ||\
		docker login $(JFROG_DOCKER_REPO_INTERNAL) --username $(DOCKER_USER) --password $(DOCKER_APIKEY)
endif
