#!/bin/bash
# Helper script to transfer a localhost podman image to a minikube docker
# instance. This works only if the pod requesting the image has the
# "imagePullPolicy: Never" or "imagePullPolicy: IfNotPresent"
# Usage: pushpodmantodocker_minikube.sh
# NOTE: Add IMG_NAME and/or IMG_TAG environment variables to change the
# default image name and tag
set -e
set -x

echo "Checking tool availability..."
which mktemp
which podman
which scp
which minikube

IMG_NAME=${IMG_NAME:-"volrep-shim-operator"}
IMG_TAG=${IMG_TAG:-"latest"}
TMP_IMG_NAME=$(mktemp)

echo "Pushing ${IMG_NAME}:${IMG_TAG} to minikube profile ${MINIKUBE_PROFILE_NAME}"

minikube profile ${MINIKUBE_PROFILE_NAME}
podman save --format docker-archive "${IMG_NAME}:${IMG_TAG}" -o "${TMP_IMG_NAME}"
scp -o 'StrictHostKeyChecking=no' -i "$(minikube ssh-key)" "${TMP_IMG_NAME}" docker@"$(minikube ip)":/tmp/
rm -f "${TMP_IMG_NAME}"
minikube ssh "docker load -i ${TMP_IMG_NAME}"
minikube ssh "docker tag localhost/${IMG_NAME}:${IMG_TAG} ${IMG_NAME}:${IMG_TAG}"
minikube ssh "rm -f ${TMP_IMG_NAME}"
minikube ssh "docker images ${IMG_NAME}:${IMG_TAG}"

echo "Success!!!"