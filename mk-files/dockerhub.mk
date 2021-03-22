.PHONY: publish-dockerhub
publish-dockerhub:
	# Dockerfile must be in same or subdirectory of this file
	docker images | grep "confluentinc/ccloud-cli" | awk '{system("docker rmi -f " "'"confluentinc/ccloud-cli"'" $2)}'
	docker build -f ./mk-files/Dockerfile_ccloud -t confluentinc/ccloud-cli:$(CLEAN_VERSION) -t confluentinc/ccloud-cli:latest ./mk-files/
	docker push confluentinc/ccloud-cli:$(CLEAN_VERSION)
	docker push confluentinc/ccloud-cli:latest
	docker images | grep "confluentinc/confluent-cli" | awk '{system("docker rmi -f " "'"confluentinc/confluent-cli"'" $2)}'
	docker build -f ./mk-files/Dockerfile_confluent -t confluentinc/confluent-cli:$(CLEAN_VERSION) -t confluentinc/confluent-cli:latest ./mk-files/
	docker push confluentinc/confluent-cli:$(CLEAN_VERSION)
	docker push confluentinc/confluent-cli:latest
