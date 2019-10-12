package memcached

import (
	"context"
	"reflect"

	cachev1alpha1 "github.com/operator-framework/operator-sdk-samples/go/memcached-operator/pkg/apis/cache/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_memcached")

// Add creates a new Memcached Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileMemcached{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("memcached-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Memcached
	err = c.Watch(&source.Kind{Type: &cachev1alpha1.Memcached{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Memcached
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &cachev1alpha1.Memcached{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &cachev1alpha1.Memcached{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileMemcached{}

// ReconcileMemcached reconciles a Memcached object
type ReconcileMemcached struct {
	// TODO: Clarify the split client
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Memcached object and makes changes based on the state read
// and what is in the Memcached.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Memcached Deployment for each Memcached CR
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileMemcached) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Memcached.")

	// Fetch the Memcached instance
	memcached := &cachev1alpha1.Memcached{}
	err := r.client.Get(context.TODO(), request.NamespacedName, memcached)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			reqLogger.Info("Memcached resource not found. Ignoring since object must be deleted.")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		reqLogger.Error(err, "Failed to get Memcached.")
		return reconcile.Result{}, err
	}

	// Update the Deployment if it already exists and was changed, if not create a new one
	// More info: https://godoc.org/sigs.k8s.io/controller-runtime/pkg/controller/controllerutil#CreateOrUpdate
	deployment := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: memcached.Name, Namespace: memcached.Namespace}}
	op, err := controllerutil.CreateOrUpdate(context.TODO(), r.client, deployment, func() error {

		// Set deployment object to expected state in memory. If it's different than what
		// the calling `CreateOrUpdate` function gets from Kubernetes, it will send an
		// update request back to the apiserver with the expected state determined here.
		deploymentForMemcached(memcached, deployment)

		// Set Memcached instance as the owner of the Deployment.
		err := controllerutil.SetControllerReference(memcached, deployment, r.scheme)
		if err != nil {
			reqLogger.Error(err, "Failed to set owner reference on memcached Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
			return err
		}
		return nil

	})
	if err != nil {
		reqLogger.Error(err, "Failed to reconcile Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
	}
	reqLogger.Info("Deployment reconciled", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name, "Operation", op)

	// Update Service if it has changed to something other than expected. If the Service doesn't
	// exist it will be created (initial creation is not really a special case).
	//
	// NOTE: The Service is used to expose the Deployment. However, the Service is not required
	// at all for the memcached example to work. The purpose is to add more examples of what you
	// can do in your operator project.
	service := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: memcached.Name, Namespace: memcached.Namespace}}
	op, err = controllerutil.CreateOrUpdate(context.TODO(), r.client, service, func() error {

		// Set service object to expected state in memory. If it's different than what the
		// calling `CreateOrUpdate` function gets from Kubernetes, it will send an update
		// request back to the apiserver with the expected state determined here.
		serviceForMemcached(memcached, service)

		// Set Memcached instance as the owner of the Service.
		err := controllerutil.SetControllerReference(memcached, service, r.scheme)
		if err != nil {
			reqLogger.Error(err, "Failed to set owner reference on memcached Service", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
			return err
		}
		return nil
	})
	if err != nil {
		reqLogger.Error(err, "Failed to reconcile Service", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
	}
	reqLogger.Info("Service reconciled", "Service.Namespace", service.Namespace, "Service.Name", service.Name, "Operation", op)

	// Update the Memcached status with the pod names
	// List the pods for this memcached's deployment
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(memcached.Namespace),
		client.MatchingLabels(labelsForMemcached(memcached.Name)),
	}
	err = r.client.List(context.TODO(), podList, listOpts...)
	if err != nil {
		reqLogger.Error(err, "Failed to list pods.", "Memcached.Namespace", memcached.Namespace, "Memcached.Name", memcached.Name)
		return reconcile.Result{}, err
	}
	podNames := getPodNames(podList.Items)

	// Update status.Nodes if needed
	if !reflect.DeepEqual(podNames, memcached.Status.Nodes) {
		memcached.Status.Nodes = podNames
		err := r.client.Status().Update(context.TODO(), memcached)
		if err != nil {
			reqLogger.Error(err, "Failed to update Memcached status.")
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

// deploymentForMemcached mutates the passed in Deployment reference to the expected state
func deploymentForMemcached(m *cachev1alpha1.Memcached, deployment *appsv1.Deployment) {
	name := "memcached"
	image := "memcached:1.4.36-alpine"
	replicas := m.Spec.Size
	ls := labelsForMemcached(m.Name)
	command := []string{"memcached", "-m=64", "-o", "modern", "-v"}
	ports := []corev1.ContainerPort{{ContainerPort: 11211, Name: "memcached", Protocol: "TCP"}}

	// Ensure all fields in the Deployment have their expected values. If they're not already
	// set to these values, we reset them here and then when control is given back to the
	// `controllerutil.CreateOrUpdate` function, it will send that update to the api server.
	if len(deployment.Spec.Template.Spec.Containers) != 1 {
		deployment.Spec.Template.Spec.Containers = make([]corev1.Container, 1)
	}

	deployment.Spec.Replicas = &replicas
	deployment.Spec.Template.ObjectMeta.Labels = ls
	deployment.Spec.Template.Spec.Containers[0].Name = name
	deployment.Spec.Template.Spec.Containers[0].Image = image
	deployment.Spec.Template.Spec.Containers[0].Ports = ports
	deployment.Spec.Template.Spec.Containers[0].Command = command

	// Deployment selector is immutable so we only set this value if
	// a new object is going to be created
	if deployment.Spec.Selector == nil {
		deployment.Spec.Selector = &metav1.LabelSelector{MatchLabels: ls}
	}

	return
}

// serviceForMemcached mutates the passed in Service reference to the expected state
func serviceForMemcached(m *cachev1alpha1.Memcached, service *corev1.Service) {

	// Ensure all fields in the Service have their expected values. If they're not already
	// set to these values, we reset them here and then when control is given back to the
	// `controllerutil.CreateOrUpdate` function, it will send that update to the api server.
	if len(service.Spec.Ports) != 1 {
		service.Spec.Ports = make([]corev1.ServicePort, 1)
	}

	service.Spec.Ports[0].Name = m.Name
	service.Spec.Ports[0].Port = 11211
	service.Spec.Ports[0].TargetPort = intstr.FromInt(11211)
	service.Spec.Ports[0].Protocol = corev1.ProtocolTCP
	service.Spec.Selector = labelsForMemcached(m.Name)

	return
}

// labelsForMemcached returns the labels for selecting the resources
// belonging to the given memcached CR name.
func labelsForMemcached(name string) map[string]string {
	return map[string]string{"app": "memcached", "memcached_cr": name}
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}
