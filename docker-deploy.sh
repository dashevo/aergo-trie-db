#! /bin/bash

# Show script in output, and error if anything fails
set -xe

GIT_TAG="$1"
if [ "x$GIT_TAG" = "x" ]; then
  echo "usage: ${0##*/} <git-tag>"
  exit 1
fi

V=$(echo "$GIT_TAG" | cut -c1)
if [ "x$V" != "xv" ]; then
  echo "error: tag should be format 'v<SemVer>', where <SemVer> is a valid semantic version string"
  exit 1
fi

TAG=$(echo "$GIT_TAG" | cut -c2-)
echo "tag=[$TAG]"

# Check semantic tag, this will fail if $TAG not valid SemVer
semvertool "$TAG"

DOCKER_TAG="latest"
PRE="$(semvertool "$TAG" --prerelease)"
if [ ! -z "$PRE" ]; then
  DOCKER_TAG="$(semvertool "$TAG" --prerelease-head)"
fi
PERMS="$(semvertool "$TAG" --show-permutations)"

# Login to Docker Hub
echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin

# Main image name / tag
IMAGE_NAME="dashpay/universe-tree-db"
docker build -t "${IMAGE_NAME}:${DOCKER_TAG}" .
docker push "${IMAGE_NAME}:${DOCKER_TAG}"

# Now do version tags
for T in $DOCKER_TAG $PERMS ; do
  docker tag "${IMAGE_NAME}:${DOCKER_TAG}" "${IMAGE_NAME}:${T}"
  docker push "${IMAGE_NAME}:${T}"
done
