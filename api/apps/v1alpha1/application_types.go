/*
Copyright 2025 Xin Yan.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ApplicationSpec defines the desired state of Application
type ApplicationSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// The following markers will use OpenAPI v3 schema to validate the value
	// More info: https://book.kubebuilder.io/reference/markers/crd-validation.html

	// Image is application docker image
	Image string `json:"image"`

	// Port is the port exposed application
	Port int32 `json:"port"`

	// Replicas refer to the desired number of identical copies (pods)
	// of an application that should be running at any given time
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// StartCmd is the application start command
	// +optional
	StartCmd string `json:"startCmd,omitempty"`

	// Args are arguments used by the application
	// +optional
	Args []string `json:"args,omitempty"`

	// Env is a list of environment variables used by the application
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`

	// Expose defines a service which exposes the application
	Expose *Expose `json:"expose"`
}

// Expose defines a service which exposes an application
type Expose struct {
	// Mode defines the service mode, NodePort or Ingress
	// +kubebuilder:validation:Enum=Ingress;NodePort
	Mode string `json:"mode"`

	// IngressDomain refers to domain name used as host in ingress
	// +optional
	IngressDomain string `json:"ingressDomain,omitempty"`

	// NodePort is a node port number for nodePort service
	// +optional
	// +kubebuilder:validation:Minimum=30000
	// +kubebuilder:validation:Maximum=32767
	NodePort int32 `json:"nodePort,omitempty"`

	// ServicePort is a port number used by the service
	// +optional
	ServicePort int32 `json:"servicePort,omitempty"`
}

// ApplicationStatus defines the observed state of Application.
type ApplicationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// For Kubernetes API conventions, see:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties

	// Deployment indicates if Deployment is ready or not ready.
	Deployment string `json:"deployment"`

	// Service indicates if Service is ready or not ready.
	Service string `json:"service"`

	// Ingress indicates if Ingress is ready or not ready.
	Ingress string `json:"ingress"`

	// Reason indicates details about why the application is in this state.
	Reason string `json:"reason"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Application is the Schema for the applications API
type Application struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of Application
	// +required
	Spec ApplicationSpec `json:"spec"`

	// status defines the observed state of Application
	// +optional
	Status ApplicationStatus `json:"status,omitempty,omitzero"`
}

// +kubebuilder:object:root=true

// ApplicationList contains a list of Application
type ApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Application `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Application{}, &ApplicationList{})
}
