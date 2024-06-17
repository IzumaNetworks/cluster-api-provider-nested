#!/bin/bash

# Boilerplate to get the directory this script is in
SOURCE=${BASH_SOURCE[0]}
while [ -L "$SOURCE" ]; do 
  DIR=$( cd -P "$( dirname "$SOURCE" )" >/dev/null 2>&1 && pwd )
  SOURCE=$(readlink "$SOURCE")
  [[ $SOURCE != /* ]] && SOURCE=$DIR/$SOURCE 
done
DIR=$( cd -P "$( dirname "$SOURCE" )" >/dev/null 2>&1 && pwd )

cd $DIR
# Function to push images
push_images() {
  for image in "${images[@]}"; do
    image_name="${image%%|*}"
    repo_location="${image#*|}"
    repo_location="${repo_location%%|*}"
    dockerfile="${image##*|}"
    echo "Pushing $repo_location:$image_name"
    docker push "$repo_location:$image_name"
  done
}

# List of tuples (image name | repo location | local dockerfile)
images=(
  "debug|ghcr.io/izumanetworks/vc-manager-amd64|Dockerfile-vc-manager-debug"
  "debug|ghcr.io/izumanetworks/vc-syncer-amd64|Dockerfile-vc-syncer-debug"
  "latest|ghcr.io/izumanetworks/vc-manager-amd64|Dockerfile-vc-manager"
  "latest|ghcr.io/izumanetworks/vc-syncer-amd64|Dockerfile-vc-syncer"
)

# Check if the -p flag is set
PUSH_IMAGES=false
while getopts ":p" opt; do
  case ${opt} in
    p ) 
      PUSH_IMAGES=true
      ;;
    \? )
      echo "Usage: $0 [-p]"
      exit 1
      ;;
  esac
done


# Build images
for image in "${images[@]}"; do
  image_name="${image%%|*}"
  repo_location="${image#*|}"
  repo_location="${repo_location%%|*}"
  dockerfile="${image##*|}"
  echo "Build $repo_location:$image_name"
  docker build -f "$dockerfile" -t "$repo_location:$image_name" .
done

# Push images if -p flag is set
if [ "$PUSH_IMAGES" = true ]; then
  push_images
fi