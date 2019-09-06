NAMESPACE=memcached

##############################
# INSTALL/UNINSTALL          #
##############################

.PHONY: install
install:
	@echo ....... Creating namespace ....... 
	- kubectl create namespace ${NAMESPACE}
	@echo ....... Applying CRDS and Operator .......
	- kubectl apply -f deploy/crds/cache_v1alpha1_memcached_crd.yaml -n ${NAMESPACE}
	@echo ....... Applying Rules and Service Account .......
	- kubectl apply -f deploy/role.yaml -n ${NAMESPACE}
	- kubectl apply -f deploy/role_binding.yaml  -n ${NAMESPACE}
	- kubectl apply -f deploy/service_account.yaml  -n ${NAMESPACE}
	@echo ....... Applying Operator .......
	- kubectl apply -f deploy/operator.yaml -n ${NAMESPACE}
	@echo ....... Creating the Database .......
	- kubectl apply -f deploy/crds/cache_v1alpha1_memcached_cr.yaml -n ${NAMESPACE}

.PHONY: uninstall
uninstall:
	@echo ....... Uninstalling .......
	@echo ....... Deleting CRDs.......
	- kubectl delete -f deploy/crds/cache_v1alpha1_memcached_crd.yaml -n ${NAMESPACE}
	@echo ....... Deleting Rules and Service Account .......
	- kubectl delete -f deploy/role.yaml -n ${NAMESPACE}
	- kubectl delete -f deploy/role_binding.yaml -n ${NAMESPACE}
	- kubectl delete -f deploy/service_account.yaml -n ${NAMESPACE}
	@echo ....... Deleting Operator .......
	- kubectl delete -f deploy/operator.yaml -n ${NAMESPACE}
	@echo ....... Deleting namespace ${NAMESPACE}.......
	- kubectl delete namespace ${NAMESPACE}

# More info: https://golang.org/cmd/vet/
.PHONY: code/vet
code/vet:
	@echo go vet
	go vet $$(go list ./... )

# More info: https://golang.org/pkg/fmt/
.PHONY: code/fmt
code/fmt:
	@echo go fmt
	go fmt $$(go list ./... )

.PHONY: code/dev
code/dev:
	@echo Running the common required commands for developments purposes
	- make code/fmt
	- make code/vet
	- make code/gen

.PHONY: code/gen
code/gen:
	@echo Updating the deep copy files with the changes in the API
	operator-sdk generate k8s
	@echo Updating the CRD files with the OpenAPI validations
	operator-sdk generate openapi

##############################
# Tests                      #
##############################

# The following command will execute many times the same e2e tests.
# Its purposes is just to illustrate few options. For further info see: https://github.com/operator-framework/operator-sdk/blob/master/doc/test-framework/writing-e2e-tests.md
.PHONY: test/e2e
test/e2e:
	@echo ... Running the same e2e tests with some few options of args ...
	@echo ... Running locally ...
	- operator-sdk test local ./test/e2e --up-local --namespace=${NAMESPACE}
	@echo ... Running NOT in parallel ...
	- operator-sdk test local ./test/e2e --go-test-flags "-v -parallel=1"
	@echo ... Running in parallel ...
	- operator-sdk test local ./test/e2e --go-test-flags "-v -parallel=2
	@echo ... Running without options/args ...
	- operator-sdk test local ./test/e2e
	@echo ... Running with the --debug param ...
	- operator-sdk test local ./test/e2e --debug
	@echo ... Running with the --verbose param ...
	- operator-sdk test local ./test/e2e --verbose
