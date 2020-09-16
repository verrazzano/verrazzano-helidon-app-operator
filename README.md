[![Go Report Card](https://goreportcard.com/badge/github.com/verrazzano/verrazzano-helidon-app-operator)](https://goreportcard.com/report/github.com/verrazzano/verrazzano-helidon-app-operator)

# verrazzano-helidon-app-operator

Kubernetes operator for handling Helidon applications

## Prerequisites

operator-sdk version must be v0.18.1

## How to Build
```
make build
```

## How to build and run outside a Kubernetes cluster

```bash
export GO111MODULE=on

make go-mod
export OPERATOR_NAME=verrazzano-helidon-app-operator
operator-sdk up local --namespace=""
```

## How to build and deploy in a Kubernetes cluster

```bash
operator-sdk build verrazzano-helidon-app-operator:v0.0.1

sed -i "" 's|REPLACE_IMAGE|verrazzano-helidon-app-operator:v0.0.1|g' deploy/operator.yaml
# This step is using the default namespace for the deployment. Change to desired
# namespace if need be.
sed -i "" "s|REPLACE_NAMESPACE|default|g" deploy/role_binding.yaml

kubectl apply -f deploy/service_account.yaml
kubectl apply -f deploy/role.yaml
kubectl apply -f deploy/role_binding.yaml
kubectl apply -f deploy/operator.yaml
```

## How to update the CRD

```bash
# Make edits to pkg/apis/verrazzano/v1beta1/helidonapp_types.go

make generate
```
