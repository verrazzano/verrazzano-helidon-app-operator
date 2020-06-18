# Copyright (c) 2020, Oracle Corporation and/or its affiliates.
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

NAME:=verrazzano-helidon-app-operator
CLUSTER_NAME = v8o-helidon-app-operator

DOCKER_IMAGE_NAME ?= ${NAME}-dev
TAG=$(shell git rev-parse HEAD)
DOCKER_IMAGE_TAG = ${TAG}

CREATE_LATEST_TAG=0

CODEGEN_PATH = k8s.io/code-generator
GOPATH ?= ${HOME}/go

ifeq ($(MAKECMDGOALS),$(filter $(MAKECMDGOALS),push push-tag))
ifndef DOCKER_REPO
    $(error DOCKER_REPO must be defined as the name of the docker repository where image will be pushed)
endif
ifndef DOCKER_NAMESPACE
    $(error DOCKER_NAMESPACE must be defined as the name of the docker namespace where image will be pushed)
endif
    DOCKER_IMAGE_FULLNAME = ${DOCKER_REPO}/${DOCKER_NAMESPACE}/${DOCKER_IMAGE_NAME}
endif

# required config for operator-sdk
export GO111MODULE=on
export OPERATOR_NAME=${NAME}

#
# Go build related tasks
#
.PHONY: go-install
go-install:
	go install ./cmd/...

.PHONY: go-fmt
go-fmt:
	gofmt -s -e -d $(shell find . -name "*.go" | grep -v /vendor/)

.PHONY: go-mod
go-mod:
	go mod vendor

	# go mod vendor only copies the .go files.  Also need
	# to populate the k8s.io/code-generator folder with the
	# scripts for generating informer/lister code

	# Obtain k8s.io/code-generator version
	$(eval codeGenVer=$(shell grep "code-generator =>" go.mod | awk '{print $$4}'))

	# Add the required files into the vendor folder
	cp ${GOPATH}/pkg/mod/${CODEGEN_PATH}@${codeGenVer}/generate-groups.sh vendor/${CODEGEN_PATH}/generate-groups.sh
	chmod +x vendor/${CODEGEN_PATH}/generate-groups.sh
	cp -R ${GOPATH}/pkg/mod/${CODEGEN_PATH}@${codeGenVer}/cmd/defaulter-gen vendor/${CODEGEN_PATH}/cmd/defaulter-gen
	chmod -R +w vendor/${CODEGEN_PATH}/cmd/defaulter-gen

#
# Build/Push-related tasks
#
.PHONY: build
build: go-mod
	# the location and name of the binary is determined (hard coded) by the opeartor-sdk - build/_output/bin/verrazzano-helidon-app-operator
	operator-sdk build ${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG}

.PHONY: generate
generate: go-mod
	./hack/update-codegen.sh
	operator-sdk generate openapi
	./hack/add-crd-header.sh

.PHONY: push
push: build
	docker tag ${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG} ${DOCKER_IMAGE_FULLNAME}:${DOCKER_IMAGE_TAG}
	docker push ${DOCKER_IMAGE_FULLNAME}:${DOCKER_IMAGE_TAG}

	if [ ${CREATE_LATEST_TAG} ]; then \
		docker tag ${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG} ${DOCKER_IMAGE_FULLNAME}:latest; \
		docker push ${DOCKER_IMAGE_FULLNAME}:latest; \
	fi


.PHONY: push-tag
push-tag:
	docker pull ${DOCKER_IMAGE_FULLNAME}:${DOCKER_IMAGE_TAG}
	docker tag ${DOCKER_IMAGE_FULLNAME}:${DOCKER_IMAGE_TAG} ${DOCKER_IMAGE_FULLNAME}:${TAG_NAME}
	docker push ${DOCKER_IMAGE_FULLNAME}:${TAG_NAME}

#
# Tests-related tasks
#
.PHONY: unit-test
unit-test: go-install
	go test -v ./pkg/apis/... ./pkg/controller/... ./cmd/...

.PHONY: coverage
coverage:
	./build/scripts/coverage.sh html

.PHONY: thirdparty-check
thirdparty-check:
	./build/scripts/thirdparty_check.sh

.PHONY: integ-test
integ-test: go-install build
	echo 'Install KinD...'
	GO111MODULE="on" go get sigs.k8s.io/kind@v0.7.0
ifdef JENKINS_URL
	./build/scripts/cleanup.sh ${CLUSTER_NAME}
endif
	echo 'Create cluster...'
	time kind create cluster \
	    --name ${CLUSTER_NAME} \
	    --wait 5m \
		--config=test/kind-config.yaml

	kubectl config set-context kind-${CLUSTER_NAME}
ifdef JENKINS_URL
	cat ${HOME}/.kube/config | grep server
	# this ugly looking line of code will get the ip address of the container running the kube apiserver
	# and update the kubeconfig file to point to that address, instead of localhost
	sed -i -e "s|127.0.0.1.*|`docker inspect ${CLUSTER_NAME}-control-plane | jq '.[].NetworkSettings.IPAddress' | sed 's/"//g'`:6443|g" ${HOME}/.kube/config
	cat ${HOME}/.kube/config | grep server
endif
	kubectl cluster-info

	kubectl wait --for=condition=ready nodes --all
	kubectl get nodes
	echo 'Copy operator Docker image into KinD...'
	kind load --name ${CLUSTER_NAME} docker-image ${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG}

	kubectl apply -f deploy/service_account.yaml
	kubectl apply -f deploy/role.yaml
	kubectl apply -f deploy/role_binding.yaml
	kubectl create -f deploy/crds/verrazzano.io_helidonapps_crd.yaml
	echo 'Deploy operator...'
	cat deploy/operator.yaml | sed -e 's|REPLACE_IMAGE|${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG}|g' | kubectl apply -f -
	echo 'Run tests...'
	go get -u github.com/onsi/ginkgo/ginkgo
	go get -u github.com/onsi/gomega/...

	kubectl get pods --all-namespaces

	ginkgo -v --keepGoing -cover test/integ/... || IGNORE=FAILURE

#
# Cleanup Kind cluster and docker containers
#
.PHONY: clean-cluster
clean-cluster:
	./build/scripts/cleanup.sh ${CLUSTER_NAME}
	
	
.PHONY: go-run
go-run: go-install
	WATCH_NAMESPACE="" go run cmd/manager/main.go --kubeconfig=${KUBECONFIG} ${EXTRA_PARAMS}

