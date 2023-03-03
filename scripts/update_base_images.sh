#!/bin/bash

eval $(gimme-aws-creds --output-format export --roles "arn:aws:iam::050879227952:role/administrator")
aws ecr get-login-password --region us-west-1 | docker login --username AWS --password-stdin 050879227952.dkr.ecr.us-west-1.amazonaws.com

push_instructions="\nTo push these images to the ECR, run the following commands:\n"

amd64_tags=($(aws ecr describe-images --registry-id 050879227952 --repository-name confluentinc/cli-centos-base-amd64 --query 'imageDetails[?imageTags.contains(@, `"latest"`)].imageTags' | jq '.[]' | jq -r '.[]'))
if [ ${#amd64_tags[@]} == 0 ]; then
  exit 1
elif [ ${#amd64_tags[@]} != 2 ]; then # check that there are exactly 2 tags for the image tagged w/ "latest"
  echo "error: there should be exactly two tags for the latest cli-centos-base-amd64 image. Current tags are: \""${amd64_tags[@]}"\"; check the ECR"
  exit 1
else
  echo "Found tags \"${amd64_tags[@]}\" for confluentinc/cli-centos-base-amd64"
fi

for tag in ${amd64_tags[@]}; do
  if [ $tag != "latest" ]; then
    integer_tag=$(printf "%.0f" $tag) && \
    incremented_tag="$(($integer_tag + 1)).0" && \
    docker build --no-cache -f ./dockerfiles/Dockerfile_linux_glibc_base -t "050879227952.dkr.ecr.us-west-1.amazonaws.com/confluentinc/cli-centos-base-amd64:$incremented_tag" . && \
    docker tag "050879227952.dkr.ecr.us-west-1.amazonaws.com/confluentinc/cli-centos-base-amd64:$incremented_tag" "050879227952.dkr.ecr.us-west-1.amazonaws.com/confluentinc/cli-centos-base-amd64:latest" && \
    push_instructions+="docker push 050879227952.dkr.ecr.us-west-1.amazonaws.com/confluentinc/cli-centos-base-amd64:$incremented_tag\n" && \
    push_instructions+="docker push 050879227952.dkr.ecr.us-west-1.amazonaws.com/confluentinc/cli-centos-base-amd64:latest\n"
  fi
done

arm64_from_amd64_tags=($(aws ecr describe-images --registry-id 050879227952 --repository-name confluentinc/cli-ubuntu-base-arm64 --query 'imageDetails[?imageTags.contains(@, `"latest"`)].imageTags' | jq '.[]' | jq -r '.[]'))
if [ ${#arm64_from_amd64_tags[@]} == 0 ]; then
  exit 1
elif [ ${#arm64_from_amd64_tags[@]} != 2 ]; then # check that there are exactly 2 tags for the image tagged w/ "latest"
  echo "error: there should be exactly two tags for the latest cli-ubuntu-base-arm64 image. Current tags are: \""${arm64_from_amd64_tags[@]}"\"; check the ECR"
  exit 1
else
  echo "Found tags \"${arm64_from_amd64_tags[@]}\" for confluentinc/cli-ubuntu-base-arm64"
fi

for tag in ${arm64_from_amd64_tags[@]}; do
  if [ $tag != "latest" ]; then
    integer_tag=$(printf "%.0f" $tag) && \
    incremented_tag="$(($integer_tag + 1)).0" && \
    docker build --no-cache -f ./dockerfiles/Dockerfile_linux_glibc_arm64_from_amd64_base -t "050879227952.dkr.ecr.us-west-1.amazonaws.com/confluentinc/cli-ubuntu-base-arm64:$incremented_tag" . && \
    docker tag "050879227952.dkr.ecr.us-west-1.amazonaws.com/confluentinc/cli-ubuntu-base-arm64:$incremented_tag" "050879227952.dkr.ecr.us-west-1.amazonaws.com/confluentinc/cli-ubuntu-base-arm64:latest" && \
    push_instructions+="docker push 050879227952.dkr.ecr.us-west-1.amazonaws.com/confluentinc/cli-ubuntu-base-arm64:$incremented_tag\n" && \
    push_instructions+="docker push 050879227952.dkr.ecr.us-west-1.amazonaws.com/confluentinc/cli-ubuntu-base-arm64:latest\n"
  fi
done

arm64_tags=($(aws ecr describe-images --registry-id 050879227952 --repository-name confluentinc/cli-ubuntu-base-arm64 --query 'imageDetails[?imageTags.contains(@, `"arm64-latest"`)].imageTags' | jq '.[]' | jq -r '.[]'))
if [ ${#arm64_tags[@]} == 0 ]; then
  exit 1
elif [ ${#arm64_tags[@]} != 2 ]; then # check that there are exactly 2 tags for the image tagged w/ "arm64-latest"
  echo "error: there should be exactly two tags for the arm64-latest cli-ubuntu-base-arm64 image. Current tags are: \""${arm64_tags[@]}"\"; check the ECR"
  exit 1
else
  echo "Found tags \"${arm64_tags[@]}\" for confluentinc/cli-ubuntu-base-arm64"
fi

for tag in ${arm64_tags[@]}; do
  if [ $tag != "arm64-latest" ]; then
    number_part=$(cut -f2 -d'-' <<< $tag) && \
    integer_tag=$(printf "%.0f" $number_part) && \
    incremented_tag="arm64-$(($integer_tag + 1)).0" && \
    docker build --no-cache -f ./dockerfiles/Dockerfile_linux_glibc_arm64_base -t "050879227952.dkr.ecr.us-west-1.amazonaws.com/confluentinc/cli-ubuntu-base-arm64:$incremented_tag" . && \
    docker tag "050879227952.dkr.ecr.us-west-1.amazonaws.com/confluentinc/cli-ubuntu-base-arm64:$incremented_tag" "050879227952.dkr.ecr.us-west-1.amazonaws.com/confluentinc/cli-ubuntu-base-arm64:arm64-latest" && \
    push_instructions+="docker push 050879227952.dkr.ecr.us-west-1.amazonaws.com/confluentinc/cli-ubuntu-base-arm64:$incremented_tag\n" && \
    push_instructions+="docker push 050879227952.dkr.ecr.us-west-1.amazonaws.com/confluentinc/cli-ubuntu-base-arm64:arm64-latest\n"
  fi
done

echo -e $push_instructions
echo -e :"\nDon't forget to commit your changes to the base Dockerfiles."