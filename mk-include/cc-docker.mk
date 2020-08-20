.PHONY: docker-login
## Login to docker Artifactory
docker-login:
ifeq ($(DOCKER_USER)$(DOCKER_APIKEY),$(_empty))
	echo "EMPTY"
	echo $(DOCKER_USER)
	@jq -e '.auths."confluent-docker.jfrog.io"' $(HOME)/.docker/config.json 2>&1 >/dev/null ||\
		(echo "confluent-docker.jfrog.io not logged in, Username and Password not found in environment, prompting for login:" && \
		 docker login confluent-docker.jfrog.io)
else
	echo "NAH EMPTY"
	@jq -e '.auths."confluent-docker.jfrog.io"' $(HOME)/.docker/config.json 2>&1 >/dev/null ||\
		docker login confluent-docker.jfrog.io --username $(DOCKER_USER) --password $(DOCKER_APIKEY)
endif
