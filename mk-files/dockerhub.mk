.PHONY: publish-dockerhub
publish-dockerhub:
	# Dockerfile must be in same or subdirectory of this file
	docker images | grep "confluentinc/confluent-cli" | awk '{system("docker rmi -f " "'"confluentinc/confluent-cli:"'" $$2)}'
	docker build --no-cache -f ./dockerfiles/Dockerfile -t confluentinc/confluent-cli:$(CLEAN_VERSION) -t confluentinc/confluent-cli:latest ./mk-files/
	$(call dry-run,docker push confluentinc/confluent-cli:$(CLEAN_VERSION))
	$(call dry-run,docker push confluentinc/confluent-cli:latest)
