#!/usr/bin/env bash

NO_COLOR=${NO_COLOR:-""}
if [ -z "$NO_COLOR" ]; then
  header_color=$'\e[1;33m'
  reset_color=$'\e[0m'
else
  header_color=''
  reset_color=''
fi

operIMG=quay.io/example/memcached-operator:v0.0.1
bundleIMG=quay.io/example-bundle/memcached-operator:v0.0.1

MEMCACHE_CONTROLLER_OLD_RECONCILER='kubebuilder'
MEMCACHE_CONTROLLER_NEW_RECONCILER='// +kubebuilder:rbac:groups=cache.example.com,resources=memcacheds,verbs=get;list;watch;create;update;patch;delete \
// +kubebuilder:rbac:groups=cache.example.com,resources=memcacheds/status,verbs=get;update;patch \
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete \
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list; \
\nfunc (r *MemcachedReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) { \
	ctx := context.Background() \
	log := r.Log.WithValues("memcached", req.NamespacedName) \
	// Fetch the Memcached instance \
	memcached := &cachev1alpha1.Memcached{} \
	err := r.Get(ctx, req.NamespacedName, memcached) \
	if err != nil { \
		if errors.IsNotFound(err) { \
			// Request object not found, could have been deleted after reconcile request. \
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers. \
			// Return and don'\''t requeue \
			log.Info("Memcached resource not found. Ignoring since object must be deleted") \
			return ctrl.Result{}, nil \
		} \
		// Error reading the object - requeue the request. \
		log.Error(err, "Failed to get Memcached") \
		return ctrl.Result{}, err \
	} \
    \n\t// Check if the deployment already exists, if not create a new one \
	found := &appsv1.Deployment{} \
	err = r.Get(ctx, types.NamespacedName{Name: memcached.Name, Namespace: memcached.Namespace}, found) \
	if err != nil && errors.IsNotFound(err) { \
		// Define a new deployment \
		dep := r.deploymentForMemcached(memcached) \
		log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name) \
		err = r.Create(ctx, dep) \
		if err != nil { \
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name) \
			return ctrl.Result{}, err \
		} \
		// Deployment created successfully - return and requeue \
		return ctrl.Result{Requeue: true}, nil \
	} else if err != nil { \
		log.Error(err, "Failed to get Deployment") \
		return ctrl.Result{}, err \
	} \
    \n\t// Ensure the deployment size is the same as the spec \
	size := memcached.Spec.Size \
	if *found.Spec.Replicas != size { \
		found.Spec.Replicas = &size \
		err = r.Update(ctx, found) \
		if err != nil { \
			log.Error(err, "Failed to update Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name) \
			return ctrl.Result{}, err \
		} \
		// Spec updated - return and requeue \
		return ctrl.Result{Requeue: true}, nil \
	} \
    \n\t// Update the Memcached status with the pod names \
	// List the pods for this memcached'\''s deployment \
	podList := &corev1.PodList{} \
	listOpts := []client.ListOption{ \
		client.InNamespace(memcached.Namespace), \
		client.MatchingLabels(labelsForMemcached(memcached.Name)), \
	} \
	if err = r.List(ctx, podList, listOpts...); err != nil { \
		log.Error(err, "Failed to list pods", "Memcached.Namespace", memcached.Namespace, "Memcached.Name", memcached.Name) \
		return ctrl.Result{}, err \
	} \
	podNames := getPodNames(podList.Items) \
    \n// Update status.Nodes if needed \
	if !reflect.DeepEqual(podNames, memcached.Status.Nodes) { \
		memcached.Status.Nodes = podNames \
		err := r.Status().Update(ctx, memcached) \
		if err != nil { \
			log.Error(err, "Failed to update Memcached status") \
			return ctrl.Result{}, err \
		} \
	} \
	\nreturn ctrl.Result{}, nil \
} \
\n\n// deploymentForMemcached returns a memcached Deployment object \
func (r *MemcachedReconciler) deploymentForMemcached(m *cachev1alpha1.Memcached) *appsv1.Deployment { \
	ls := labelsForMemcached(m.Name) \
	replicas := m.Spec.Size \
	\ndep := &appsv1.Deployment{ \
		ObjectMeta: metav1.ObjectMeta{ \
			Name:      m.Name, \
			Namespace: m.Namespace, \
		}, \
		Spec: appsv1.DeploymentSpec{ \
			Replicas: &replicas, \
			Selector: &metav1.LabelSelector{ \
				MatchLabels: ls, \
			}, \
			Template: corev1.PodTemplateSpec{ \
				ObjectMeta: metav1.ObjectMeta{ \
					Labels: ls, \
				}, \
				Spec: corev1.PodSpec{ \
					Containers: []corev1.Container{{ \
						Image:   "memcached:1.4.36-alpine", \
						Name:    "memcached", \
						Command: []string{"memcached", "-m=64", "-o", "modern", "-v"}, \
						Ports: []corev1.ContainerPort{{ \
							ContainerPort: 11211, \
							Name:          "memcached", \
						}}, \
					}}, \
				}, \
			}, \
		}, \
	} \
	// Set Memcached instance as the owner and controller \
	ctrl.SetControllerReference(m, dep, r.Scheme) \
	return dep \
} \
\n// labelsForMemcached returns the labels for selecting the resources \
// belonging to the given memcached CR name. \
func labelsForMemcached(name string) map[string]string { \
	return map[string]string{"app": "memcached", "memcached_cr": name} \
} \
\n// getPodNames returns the pod names of the array of pods passed in \
func getPodNames(pods []corev1.Pod) []string { \
	var podNames []string \
	for _, pod := range pods { \
		podNames = append(podNames, pod.Name) \
	} \
	return podNames \
} \
\nfunc (r *MemcachedReconciler) SetupWithManager(mgr ctrl.Manager) error { \
	return ctrl.NewControllerManagedBy(mgr). \
		For(&cachev1alpha1.Memcached{}). \
		Owns(&appsv1.Deployment{}). \
		Complete(r) \
}'

MAKEFILE_UNDEPLOY_TARGET='$a \
\n# UnDeploy controller from the configured Kubernetes cluster in ~/.kube/config \
# Note that it was added for we are allowed to uninstall the files. However, it will be present by default in the future \
# versions. \
undeploy: \
	$(KUSTOMIZE) build config\/default | kubectl delete -f -'

MAKEFILE_OLD_TEST_TARGET='go test .\/... -coverprofile cover.out'

MAKEFILE_NEW_TEST_TARGET='mkdir -p ${ENVTEST_ASSETS_DIR} \
	test -f ${ENVTEST_ASSETS_DIR}\/setup-envtest.sh || curl -sSLo ${ENVTEST_ASSETS_DIR}\/setup-envtest.sh https:\/\/raw.githubusercontent.com\/kubernetes-sigs\/controller-runtime\/master\/hack\/setup-envtest.sh \
	source ${ENVTEST_ASSETS_DIR}\/setup-envtest.sh; fetch_envtest_tools $(ENVTEST_ASSETS_DIR); setup_envtest_env $(ENVTEST_ASSETS_DIR); go test .\/... -coverprofile cover.out'

MAKEFILE_PACKAGE_MANIFEST_TARGET='$a \
\n \
# Options for "packagemanifests". \
ifneq ($(origin CHANNEL), undefined) \
PKG_CHANNELS := --channel=$(CHANNEL) \
endif \
ifeq ($(IS_CHANNEL_DEFAULT), 1) \
PKG_IS_DEFAULT_CHANNEL := --default-channel \
endif \
PKG_MAN_OPTS ?= $(PKG_CHANNELS) $(PKG_IS_DEFAULT_CHANNEL) \
# Generate package manifests. \
packagemanifests: kustomize manifests \
	operator-sdk generate kustomize manifests -q --interactive=false \
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(IMG) \
	$(KUSTOMIZE) build config\/manifests | operator-sdk generate packagemanifests -q --version $(VERSION) $(PKG_MAN_OPTS)'

function header_text {
  echo "$header_color$*$reset_color"
}

function update_memcached_types() {
    FILENAME=api/v1alpha1/memcached_types.go
    
    sed -i 's@Foo string `json:"foo,omitempty"`@\
    Size int32 `json:"size"`@' $FILENAME

    sed -i '/type MemcachedStatus struct {/ a\
    Nodes []string `json:"nodes"`' $FILENAME

    sed -i '/\/\/ Foo is an example field of Memcached. Edit Memcached_types.go to remove\/update/d' $FILENAME
}

function update_memcached_controllers() {
    FILENAME=controllers/memcached_controller.go
    
    # Updating imports
    sed -i '/context/ a\
    "reflect"' $FILENAME

    sed -i '/github.com\/go-logr\/logr/ a\
    appsv1 "k8s.io/api/apps/v1" \
    corev1 "k8s.io/api/core/v1" \
    "k8s.io/apimachinery/pkg/api/errors" \
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"' $FILENAME

    sed -i '/apimachinery\/pkg\/runtime/ a\
    "k8s.io/apimachinery/pkg/types"' $FILENAME

    sed -i '/'"$MEMCACHE_CONTROLLER_OLD_RECONCILER"'/,$c\'"$MEMCACHE_CONTROLLER_NEW_RECONCILER" $FILENAME
}

function update_memcached_webhook() {
    FILENAME=api/v1alpha1/memcached_webhook.go

    # Update imports
    sed -i '/import (/ a\
    "errors"' $FILENAME

    # Updating validation logic
    sed -i 's/\/\/ TODO(user): fill in your defaulting logic./ if r.Spec.Size == 0 { \
		r.Spec.Size = 3 \
	}/' $FILENAME

    sed -i '/return nil/d' $FILENAME

    sed -i 's/\/\/ TODO(user): fill in your validation logic upon object creation./ return validateOdd(r.Spec.Size)/' $FILENAME

    sed -i 's/\/\/ TODO(user): fill in your validation logic upon object update./ return validateOdd(r.Spec.Size)/' $FILENAME

    sed -i 's/\/\/ TODO(user): fill in your validation logic upon object deletion./ return nil/' $FILENAME


    sed -i -e '$afunc validateOdd(n int32) error { \
	    if n%2 == 0 { \
		    return errors.New("Cluster size must be an odd number") \
	    } \
	    return nil \
    }' $FILENAME
}

function update_default_kustomization() {
    FILENAME=config/default/kustomization.yaml 

    # Uncommenting lines with [WEBHOOK] prefix in 
    # config/default/kustomization.yaml to enable Webhooks
    sed -i '45,70 s/^#//' $FILENAME
    sed -i '21 s/^#//' $FILENAME
    sed -i '23 s/^#//' $FILENAME
    sed -i '35 s/^#//' $FILENAME
    sed -i '40 s/^#//' $FILENAME
}

function update_cache_v1alpha1_memcached_sample() {
    sed -i 's/foo: bar/size: 3/' config/samples/cache_v1alpha1_memcached.yaml
}

function update_Makefile() {
	FILENAME=Makefile

	# Add undeploy target
	sed -i "$MAKEFILE_UNDEPLOY_TARGET" $FILENAME

	# Modify the test target
	sed -i 's/'"$MAKEFILE_OLD_TEST_TARGET"'/'"$MAKEFILE_NEW_TEST_TARGET"/ $FILENAME

	# Add Envtest variable
	sed -i '/# Run tests/a ENVTEST_ASSETS_DIR=$(shell pwd)\/testbin' $FILENAME

	# Add packagemanifest target
	sed -i "$MAKEFILE_PACKAGE_MANIFEST_TARGET" $FILENAME
}

function scaffold_go_project() {

    header_text "Removing the existing example project"
    rm -rf memcached-operator

    header_text "Creating project directory"
    mkdir memcached-operator
    cd memcached-operator

    header_text "Initializing new project and creating API"
    operator-sdk init --repo github.com/example/memcached-operator --domain example.com
    operator-sdk create api --group cache --version v1alpha1 --kind Memcached --controller --resource

    # Update memcached_types.go
    update_memcached_types

    # Update memcached_controllers.go
    update_memcached_controllers

    header_text "Updating the generated code for specific resource"
    make generate

    header_text "Generating manifests"
    make manifests

    header_text "Scaffolding webhook"
    operator-sdk create webhook --group cache --version v1alpha1 --kind Memcached --defaulting --programmatic-validation

    # Update memcached_webhook.go
    update_memcached_webhook

    # Update config/default/kustomization.yaml
    update_default_kustomization

    # Update config/samples/cache_v1alpha1_memcached.yaml
    update_cache_v1alpha1_memcached_sample

	# Add Undeploy target to Makefile
	update_Makefile

	header_text "integrating with OLM ..."
	sed -i".bak" -E -e 's/operator-sdk generate kustomize manifests/operator-sdk generate kustomize manifests --interactive=false/g' Makefile; rm -f Makefile.bak

	# OLM Integration
	operIMG=quay.io/example/memcached-operator:v0.0.1
	bundleIMG=quay.io/example-bundle/memcached-operator:v0.0.1
	
	header_text "generating bundle and building the image ..."
	make bundle IMG=$operIMG
	make bundle-build BUNDLE_IMG=$bundleIMG

	make packagemanifests
}

scaffold_go_project

# Run go fmt against code
go fmt ./...
