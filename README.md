[![pipeline status](https://github.com/verrazzano/verrazzano-helidon-app-operator/badges/master/pipeline.svg)](https://github.com/verrazzano/verrazzano-helidon-app-operator/commits/master)
[![coverage report](https://github.com/verrazzano/verrazzano-helidon-app-operator/badges/master/coverage.svg)](https://github.com/verrazzano/verrazzano-helidon-app-operator/commits/master)

# verrazzano-helidon-app-operator

Kubernetes operator for handling Helidon applications

## Prerequisites

operator-sdk version must be v0.12.0

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

## Generating/Updating THIRD_PARTY_LICENSES.txt

Whenever project dependencies (go.mod) are updated, the `THIRD_PARTY_LICENSES.txt` file contained in this project must be updated as well.
This is verified in the CI pipeline - the build will fail if this file is found to be out of sync with
go.mod.

To update the `THIRD_PARTY_LICENSES.txt` file, install the *Attribution Helper* tool as described [here](https://github.com/oracle/attribution-helper#how-to-use),
run it within this project's the root directory:

```
attribution-helper gen
```

and then commit the updated `THIRD_PARTY_LICENSES.txt` file.
