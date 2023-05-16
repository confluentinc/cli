#!/bin/bash

export AWS_PROFILE=cc-production-1/prod-administrator
aws ecr get-login-password --region us-west-1 | docker login --username AWS --password-stdin 050879227952.dkr.ecr.us-west-1.amazonaws.com

new_tags="\nThe following tags ready to be pushed to ECR:\n"

arm64_from_amd64_tags=($(aws ecr describe-images --registry-id 050879227952 --repository-name confluentinc/cli-ubuntu-base-arm64 --query 'imageDetails[?imageTags.contains(@, `"latest"`)].imageTags' | jq '.[]' | jq -r '.[]'))
if [ ${#arm64_from_amd64_tags[@]} == 0 ]; then
  exit 1
elif [ ${#arm64_from_amd64_tags[@]} != 2 ]; then # check that there are exactly 2 tags for the image tagged w/ "latest"
  echo "error: there should be exactly two tags for the latest cli-ubuntu-base-arm64 image. Current tags are: \""${arm64_from_amd64_tags[@]}"\"; check the ECR"
  exit 1
else
  echo "Found tags \"${arm64_from_amd64_tags[@]}\" for confluentinc/cli-ubuntu-base-arm64"
fi

incremented_tag=""
for tag in ${arm64_from_amd64_tags[@]}; do
  if [ $tag != "latest" ]; then
    integer_tag=$(printf "%.0f" $tag) && \
    incremented_tag+="$(($integer_tag + 1)).0" && \
    docker build --no-cache -f ./docker/Dockerfile_linux_glibc_arm64_from_amd64_base -t "050879227952.dkr.ecr.us-west-1.amazonaws.com/confluentinc/cli-ubuntu-base-arm64:$incremented_tag" . && \
    docker tag "050879227952.dkr.ecr.us-west-1.amazonaws.com/confluentinc/cli-ubuntu-base-arm64:$incremented_tag" "050879227952.dkr.ecr.us-west-1.amazonaws.com/confluentinc/cli-ubuntu-base-arm64:latest" && \
    new_tags+="050879227952.dkr.ecr.us-west-1.amazonaws.com/confluentinc/cli-ubuntu-base-arm64:$incremented_tag\n" && \
    new_tags+="050879227952.dkr.ecr.us-west-1.amazonaws.com/confluentinc/cli-ubuntu-base-arm64:latest\n"
  fi
done

echo -e $new_tags

read -p "Push these tags to ECR? (y/n)"
if [ $REPLY == "y" ]; then
  docker push "050879227952.dkr.ecr.us-west-1.amazonaws.com/confluentinc/cli-ubuntu-base-arm64:$incremented_tag"
  docker push "050879227952.dkr.ecr.us-west-1.amazonaws.com/confluentinc/cli-ubuntu-base-arm64:latest"
fi
