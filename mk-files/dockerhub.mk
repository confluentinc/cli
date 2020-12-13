.PHONY: publish-dockerhub
publish-dockerhub:
	# Dockerfile must be in same or subdirectory of this file
	docker build -f ./mk-files/Dockerfile_ccloud -t confluentinc/ccloud-cli:$(CLEAN_VERSION) ./mk-files/
	docker push confluentinc/ccloud-cli

