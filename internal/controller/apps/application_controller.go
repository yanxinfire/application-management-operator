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

package apps

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	appsv1alpha1 "github.com/yanxinfire/application-management-operator/api/apps/v1alpha1"
)

// ApplicationReconciler reconciles an Application object
type ApplicationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=apps.xinyan.cn,resources=applications,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.xinyan.cn,resources=applications/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps.xinyan.cn,resources=applications/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Application object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/reconcile
func (r *ApplicationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logf.FromContext(ctx, "Application", req.NamespacedName)

	app := &appsv1alpha1.Application{}
	if err := r.Get(ctx, req.NamespacedName, app); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	appCopy := app.DeepCopy()
	deploy := &appsv1.Deployment{}
	if err := r.Get(ctx, req.NamespacedName, deploy); err != nil {
		if errors.IsNotFound(err) {
			r.createDeployment()
		} else {
			logger.Error(err, "Deployment not found")
			return ctrl.Result{}, err
		}
	} else {
		r.updateDeployment()
	}

	if app.Spec.Expose.Mode == "Ingress" {
		if err := r.createOrUpdateIngress(ctx, req); err != nil {
			return ctrl.Result{}, err
		}
	} else if app.Spec.Expose.Mode == "NodePort" {
		r.createOrUpdateNodePortService()
	} else {
		return ctrl.Result{}, fmt.Errorf("expose mode %s is not supported", app.Spec.Expose.Mode)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.Application{}).
		Named("apps-application").
		Complete(r)
}

func (r *ApplicationReconciler) createDeployment() {

}

func (r *ApplicationReconciler) updateDeployment() {

}

func (r *ApplicationReconciler) createOrUpdateIngress(ctx context.Context, req ctrl.Request, expose appsv1alpha1.Expose) error {
	ing := &networkingv1.Ingress{}
	if err := r.Get(ctx, req.NamespacedName, ing); err != nil {
		if errors.IsNotFound(err) {
			return r.createIngress(ctx, req, expose)
		} else {
			return err
		}
	}
	return r.updateIngress(ctx, req, ing.DeepCopy())
}

func (r *ApplicationReconciler) createOrUpdateNodePortService() {

}

func (r *ApplicationReconciler) createIngress(ctx context.Context, req ctrl.Request, expose appsv1alpha1.Expose) error {

	svc := &corev1.Service{}
	if err := r.Get(ctx, req.NamespacedName, svc); err != nil {
		if errors.IsNotFound(err) {
			return r.createService(ctx, req, expose)
		} else {
			return err
		}
	}
	ing := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: req.Namespace,
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: expose.IngressDomain,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path: "/",
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: req.Name,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	return r.Create(ctx, ing)
}

func (r *ApplicationReconciler) updateIngress(ctx context.Context, req ctrl.Request, deepCopy *networkingv1.Ingress) error {
	ing := networkingv1.Ingress{}
	err := r.Create(ctx, &ing)
	if err != nil {
		return err
	}
}

func (r *ApplicationReconciler) createService(ctx context.Context, req ctrl.Request, expose appsv1alpha1.Expose) error {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: req.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name: "http",
					Port: expose.ServicePort,
				},
			},
		},
	}
	if expose.Mode == "NodePort" {
		svc.Spec.Ports[0].NodePort = expose.NodePort
	}
	return r.Create(ctx, svc)
}
