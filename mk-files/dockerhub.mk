.PHONY: publish-dockerhub
publish-dockerhub:
	# Dockerfile must be in same or subdirectory of this file
	docker build -f ./mk-files/Dockerfile_ccloud -t confluentinc/ccloud-cli:$(CLEAN_VERSION) -t confluentinc/ccloud-cli:latest ./mk-files/
	docker push --all-tags confluentinc/ccloud-cli
	docker build -f ./mk-files/Dockerfile_confluent -t confluentinc/confluent-cli:$(CLEAN_VERSION) -t confluentinc/confluent-cli:latest ./mk-files/
	docker push --all-tags confluentinc/confluent-cli
