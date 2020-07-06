// Copyright (c) 2020, Oracle Corporation and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HelidonAppSpec defines the desired state of HelidonApp
// +k8s:openapi-gen=true
type HelidonAppSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html

	// User defined description of the the HelidonApp custom resource
	Description string `json:"description"`
	// The name of the Helidon application
	Name string `json:"name"`
	// The namespace for the Helidon application
	Namespace string `json:"namespace"`
	// The docker image to pull
	Image string `json:"image"`
	// The Kubernetes docker secrets for pulling images
	// +x-kubernetes-list-type=set
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	// The Kubernetes pull policy for pulling the image
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
	// The Kubernetes ServiceAccount name to run this pod
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
	// Number of replicas to create.
	// This is a pointer to distinguish between explicit zero and not specified.
	// Defaults to 1.
	Replicas *int32 `json:"replicas,omitempty"`
	// Port to be used for service - defaults to 8080
	Port int32 `json:"port,omitempty"`
	// Port to be used for service targetPort - defaults to 8080
	TargetPort int32 `json:"targetPort,omitempty"`
	// Array of environment variables for image
	// +x-kubernetes-list-type=set
	Env []corev1.EnvVar `json:"env,omitempty"`
	// InitContainers holds a list of initialization containers that should
	// be run before starting the main container in this pod.
	// +x-kubernetes-list-type=set
	InitContainers []corev1.Container `json:"initContainers,omitempty"`
	// Containers to be included in the pod
	// +x-kubernetes-list-type=set
	Containers []corev1.Container `json:"containers,omitempty"`
	// Volumes to be created in the pod
	// +x-kubernetes-list-type=set
	Volumes []corev1.Volume `json:"volumes,omitempty"`
}

// HelidonAppStatus defines the observed state of HelidonApp
// +k8s:openapi-gen=true
type HelidonAppStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html

	// State of the Helidon deployment
	State string `json:"state,omitempty"`
	// Message associated with latest action
	LastActionMessage string `json:"lastActionMessage,omitempty"`
	// Time stamp for latest action
	LastActionTime string `json:"lastActionTime,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// HelidonApp is the Schema for the helidonapps API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=ha
// +genclient
// +genclient:noStatus
type HelidonApp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HelidonAppSpec   `json:"spec,omitempty"`
	Status HelidonAppStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// HelidonAppList contains a list of HelidonApp
type HelidonAppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HelidonApp `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HelidonApp{}, &HelidonAppList{})
}
