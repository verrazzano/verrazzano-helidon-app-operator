// Copyright (c) 2020, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package helidonapp

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"go.uber.org/zap"

	verrazzanov1beta1 "github.com/verrazzano/verrazzano-helidon-app-operator/pkg/apis/verrazzano/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new HelidonApp Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileHelidonApp{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("helidonapp-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource HelidonApp
	err = c.Watch(&source.Kind{Type: &verrazzanov1beta1.HelidonApp{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileHelidonApp implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileHelidonApp{}

// ReconcileHelidonApp reconciles a HelidonApp object
type ReconcileHelidonApp struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a HelidonApp object and makes changes based on the state read
// and what is in the HelidonApp.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileHelidonApp) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// create logger with initialized values for reconciliation
	reqLogger := zap.S().With("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Infow("Reconciling HelidonApp")
	// Fetch the HelidonApp instance
	instance := &verrazzanov1beta1.HelidonApp{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		// If the resource is not found, that means all of
		// the finalizers have been removed, and the CohCluster
		// resource has been deleted, so there is nothing left
		// to do.
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}

		// Error reading the object - requeue the request.
		reqLogger.Errorf("Failed to get HelidonApp, Error: %s", err.Error())
		return reconcile.Result{}, err
	}

	// Check if the namespace for the Helidon application exists, if not found create it
	// Define a new Namespace object
	namespaceFound := &corev1.Namespace{}
	reqLogger.Infow("Checking if namespace exist")
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Spec.Namespace}, namespaceFound)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Infof("Creating a new namespace, Namespace: %s", instance.Spec.Namespace)
		err = r.client.Create(context.TODO(), newNamespace(instance))
		if err != nil {
			return reconcile.Result{}, err
		}

		// Namespace created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Check if the serviceaccount for the Helidon application exists, if not found create it
	// Define a new ServiceAccount object
	if instance.Spec.ServiceAccountName != "" {
		saFound := &corev1.ServiceAccount{}
		reqLogger.Infow("Checking if serviceaccount exist")
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Spec.ServiceAccountName, Namespace: instance.Spec.Namespace}, saFound)
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Infof("Creating a new serviceaccount, Name: %s Namespace: %s", instance.Spec.ServiceAccountName, instance.Spec.Namespace)
			err = r.client.Create(context.TODO(), newServiceAccount(instance))
			if err != nil {
				return reconcile.Result{}, err
			}

			// serviceaccount created successfully - return and requeue
			return reconcile.Result{Requeue: true}, nil
		} else if err != nil {
			return reconcile.Result{}, err
		}
	}

	// Define a new Deployment object
	deployment := newDeployment(instance)

	// Set HelidonApp instance as the owner and controller of the deployment
	// This reference will result in the deployment resource being deleted when the CR is deleted
	if err := controllerutil.SetControllerReference(instance, deployment, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Deployment already exists
	deployFound := &appsv1.Deployment{}
	reqLogger.Infow("Checking if deployment exist")
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: deployment.Name, Namespace: deployment.Namespace}, deployFound)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Infof("Creating a new Deployment. Name: %s Namespace: %s", deployment.Name, deployment.Namespace)
		err = r.client.Create(context.TODO(), deployment)
		if err != nil {
			reqLogger.Errorf("Failed to create Deployment, Name: %s Namespace: %s, Error: %s", deployment.Name, deployment.Namespace, err.Error())
			r.updateStatus(reqLogger, instance, "Failed", "Helidon application deployment creation failed: "+err.Error())
			return reconcile.Result{}, err
		}

		r.updateStatus(reqLogger, instance, "Deployed", "Helidon application deployed successfully")

		// Deployment created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Define a new Service object
	service := newService(instance)

	// Set HelidonApp instance as the owner and controller of the service
	// This reference will result in the service resource being deleted when the CR is deleted
	if err := controllerutil.SetControllerReference(instance, service, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Service already exists
	serviceFound := &corev1.Service{}
	reqLogger.Infow("Checking if service exist")
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, serviceFound)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Infof("Creating a new Service, Name: %s Namespace %s", service.Name, service.Namespace)
		err = r.client.Create(context.TODO(), service)
		if err != nil {
			reqLogger.Errorf("Failed to create Service, Name: %s Namespace: %s, Error: %s", service.Name, service.Namespace, err.Error())
			r.updateStatus(reqLogger, instance, "Failed", "Helidon application service creation failed: "+err.Error())
			return reconcile.Result{}, err
		}

		r.updateStatus(reqLogger, instance, "Deployed", "Helidon application service deployed successfully")

		// Service created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Let's update the deployment if needed
	err = r.doUpdateIfNeeded(reqLogger, instance, deployFound)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Let's update the service if needed
	if instance.Spec.Port != 0 && instance.Spec.Port != serviceFound.Spec.Ports[0].Port {
		reqLogger.Infof("Updating Service, Name: %s Namespace: %s", serviceFound.Name, serviceFound.Namespace)
		serviceFound.Spec.Ports[0].Port = instance.Spec.Port
		serviceFound.Spec.Ports[0].TargetPort = intstr.IntOrString{
			Type:   intstr.Int,
			IntVal: instance.Spec.Port,
		}
		err = r.client.Update(context.TODO(), serviceFound)
		if err != nil {
			reqLogger.Errorf("Failed to update Service, Name: %s Namespace: %s, Error: %s", service.Name, service.Namespace, err.Error())
			r.updateStatus(reqLogger, instance, instance.Status.State, "Helidon application service update failed: "+err.Error())
			return reconcile.Result{}, err
		}

		r.updateStatus(reqLogger, instance, "Updated", "Helidon application service updated")
	}

	// Helidon application updated - don't requue
	return reconcile.Result{}, nil
}

// newDeployment returns a deployment for creating/updating a Helidon application deployment
func newDeployment(cr *verrazzanov1beta1.HelidonApp) *appsv1.Deployment {
	labels := make(map[string]string)
	labels["app"] = cr.Spec.Name

	port, targetPort := getPorts(cr)

	annotations := make(map[string]string)
	annotations["prometheus.io/scrape"] = "true"
	annotations["prometheus.io/port"] = fmt.Sprint(targetPort)
	annotations["prometheus.io/path"] = "/metrics"

	containers := []corev1.Container{
		{
			Name:            cr.Spec.Name,
			Image:           cr.Spec.Image,
			ImagePullPolicy: cr.Spec.ImagePullPolicy,
			Ports: []corev1.ContainerPort{
				{
					ContainerPort: port,
				},
			},
			Env: cr.Spec.Env,
		},
	}

	// Include any additional containers specified in the CR
	for _, container := range cr.Spec.Containers {
		containers = append(containers, container)
	}

	return &appsv1.Deployment{

		ObjectMeta: metav1.ObjectMeta{
			Name:        cr.Spec.Name,
			Namespace:   cr.Spec.Namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: func() *int32 {
				if cr.Spec.Replicas != nil {
					return cr.Spec.Replicas
				}
				// Return default of 1 if not specified
				var val int32 = 1
				return &val
			}(),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
					Annotations: annotations,
				},
				Spec: corev1.PodSpec{
					InitContainers:     cr.Spec.InitContainers,
					Containers:         containers,
					ServiceAccountName: cr.Spec.ServiceAccountName,
					ImagePullSecrets:   cr.Spec.ImagePullSecrets,
					Volumes:            cr.Spec.Volumes,
				},
			},
		},
	}
}

// newService returns a service for creating a Helidon application service
func newService(cr *verrazzanov1beta1.HelidonApp) *corev1.Service {
	labels := make(map[string]string)
	labels["app"] = cr.Spec.Name

	port, targetPort := getPorts(cr)

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Spec.Name,
			Namespace: cr.Spec.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Name: "http",
					Port: port,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: targetPort,
					},
				},
			},
		},
	}
}

// Get the port and targetPort values
func getPorts(cr *verrazzanov1beta1.HelidonApp) (int32, int32) {
	// Default port value is 8080
	var port int32 = 8080
	if cr.Spec.Port != 0 {
		port = cr.Spec.Port
	}
	// Default target port value is value of port
	var targetPort = port
	if cr.Spec.TargetPort != 0 {
		targetPort = cr.Spec.TargetPort
	}

	return port, targetPort
}

// createNamespace returns a namespace resource that may need to be created
func newNamespace(cr *verrazzanov1beta1.HelidonApp) *corev1.Namespace {
	labels := make(map[string]string)
	labels["istio-injection"] = "enabled"

	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   cr.Spec.Namespace,
			Labels: labels,
		},
	}

	return namespace
}

// createServiceAccount returns a serviceaccount resource that may need to be created
func newServiceAccount(cr *verrazzanov1beta1.HelidonApp) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Spec.ServiceAccountName,
			Namespace: cr.Spec.Namespace,
		},
	}
}

// doUpdateIfNeeded does an update if needed
func (r *ReconcileHelidonApp) doUpdateIfNeeded(reqLogger *zap.SugaredLogger, cr *verrazzanov1beta1.HelidonApp, deployFound *appsv1.Deployment) error {
	updateNeeded := false
	if !isReplicasEqual(deployFound.Spec.Replicas, cr.Spec.Replicas) {
		if cr.Spec.Replicas != nil {
			deployFound.Spec.Replicas = cr.Spec.Replicas
		} else {
			// default of 1 if not specified
			var val int32 = 1
			deployFound.Spec.Replicas = &val
		}
		updateNeeded = true
	}
	if deployFound.Spec.Template.Spec.Containers[0].Image != cr.Spec.Image {
		deployFound.Spec.Template.Spec.Containers[0].Image = cr.Spec.Image
		updateNeeded = true
	}
	if cr.Spec.Port != 0 && deployFound.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort != cr.Spec.Port {
		deployFound.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort = cr.Spec.Port
		updateNeeded = true
	}
	if !reflect.DeepEqual(deployFound.Spec.Template.Spec.Containers[0].Env, cr.Spec.Env) {
		deployFound.Spec.Template.Spec.Containers[0].Env = cr.Spec.Env
		updateNeeded = true
	}
	if !reflect.DeepEqual(deployFound.Spec.Template.Spec.ImagePullSecrets, cr.Spec.ImagePullSecrets) {
		deployFound.Spec.Template.Spec.ImagePullSecrets = cr.Spec.ImagePullSecrets
		updateNeeded = true
	}
	if deployFound.Spec.Template.Spec.ServiceAccountName != cr.Spec.ServiceAccountName {
		deployFound.Spec.Template.Spec.ServiceAccountName = cr.Spec.ServiceAccountName
		updateNeeded = true
	}
	if !reflect.DeepEqual(deployFound.Spec.Template.Spec.InitContainers, cr.Spec.InitContainers) {
		deployFound.Spec.Template.Spec.InitContainers = cr.Spec.InitContainers
		updateNeeded = true
	}
	if len(cr.Spec.Volumes) != len(deployFound.Spec.Template.Spec.Volumes) {
		var volumes []corev1.Volume
		for _, crVolume := range cr.Spec.Volumes {
			volumes = append(volumes, crVolume)
		}
		deployFound.Spec.Template.Spec.Volumes = volumes
		updateNeeded = true
	}
	if len(cr.Spec.Containers)+1 != len(deployFound.Spec.Template.Spec.Containers) {
		var containers []corev1.Container
		containers = append(containers, deployFound.Spec.Template.Spec.Containers[0])
		for _, crContainer := range cr.Spec.Containers {
			containers = append(containers, crContainer)
		}
		deployFound.Spec.Template.Spec.Containers = containers
		updateNeeded = true
	}

	if updateNeeded {
		reqLogger.Infof("Updating Deployment, Name: %s Namespace: %s", deployFound.Name, deployFound.Namespace)
		err := r.client.Update(context.TODO(), deployFound)
		if err != nil {
			reqLogger.Errorf("Failed to update Deployment, Name: %s Namespace: %s, Error: %s", deployFound.Name, deployFound.Namespace, err.Error())
			r.updateStatus(reqLogger, cr, cr.Status.State, "Helidon application deployment update failed: "+err.Error())
			return err
		}

		r.updateStatus(reqLogger, cr, "Updated", "Helidon application deployment updated")
	}

	return nil
}

// Check if the existing and target replicas are same
func isReplicasEqual(existing *int32, target *int32) bool {
	if existing == nil && target == nil {
		return true
	}

	if existing == nil && target != nil {
		return false
	}

	if existing != nil && target == nil {
		if *existing == 1 {
			return true
		}
		return false
	}

	if *existing == *target {
		return true
	}

	return false
}

// Update the status for the CR
func (r *ReconcileHelidonApp) updateStatus(reqLogger *zap.SugaredLogger, cr *verrazzanov1beta1.HelidonApp, state string, message string) error {
	cr.Status.State = state
	cr.Status.LastActionMessage = message
	t := time.Now().UTC()
	cr.Status.LastActionTime = fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02dZ",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())

	// Update status in CR
	err := r.client.Status().Update(context.TODO(), cr)
	if err != nil {
		reqLogger.Errorf("Failed to update Helidon application status, Error: %s", err.Error())
		return err
	}
	return nil

}
