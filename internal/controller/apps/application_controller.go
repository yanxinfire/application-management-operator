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
	"time"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	appsv1alpha1 "github.com/yanxinfire/application-management-operator/api/apps/v1alpha1"
)

// ApplicationReconciler reconciles an Application object
type ApplicationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	logger logr.Logger
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
	r.logger = logf.FromContext(ctx, "Application", req.NamespacedName)

	app := &appsv1alpha1.Application{}
	if err := r.Get(ctx, req.NamespacedName, app); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	if err := r.verifyApplicationMode(app); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.createOrUpdateDeployment(ctx, app); err != nil {
		return ctrl.Result{RequeueAfter: 30 * time.Second}, err
	}

	if err := r.createOrUpdateService(ctx, app); err != nil {
		return ctrl.Result{RequeueAfter: 30 * time.Second}, err
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

func (r *ApplicationReconciler) createOrUpdateDeployment(
	ctx context.Context, app *appsv1alpha1.Application) error {
	deployment := NewDeployment(app)
	if err := r.Get(ctx, types.NamespacedName{
		Namespace: app.Namespace,
		Name:      app.Name,
	}, &appsv1.Deployment{}); err != nil {
		if errors.IsNotFound(err) {
			return r.Create(ctx, deployment)
		}
		return err
	}
	return r.Update(ctx, deployment)
}

func (r *ApplicationReconciler) createOrUpdateService(
	ctx context.Context, app *appsv1alpha1.Application) error {
	service := NewService(app)
	if app.Spec.Expose.Mode == "Ingress" {
		return r.createOrUpdateIngress(ctx, app)
	}
	if err := r.Get(ctx, types.NamespacedName{
		Namespace: app.Namespace,
		Name:      app.Name,
	}, &corev1.Service{}); err != nil {
		if errors.IsNotFound(err) {
			return r.Create(ctx, service)
		}
		return err
	}
	return r.Update(ctx, service)
}

func (r *ApplicationReconciler) createOrUpdateIngress(
	ctx context.Context, app *appsv1alpha1.Application) error {
	ingress := NewService(app)
	if err := r.Get(ctx, types.NamespacedName{
		Namespace: app.Namespace,
		Name:      app.Name,
	}, &networkingv1.Ingress{}); err != nil {
		if errors.IsNotFound(err) {
			return r.Create(ctx, ingress)
		}
		return err
	}
	return r.Update(ctx, ingress)
}

func (r *ApplicationReconciler) verifyApplicationMode(app *appsv1alpha1.Application) error {
	expose := app.Spec.Expose
	if expose.Mode == "Ingress" {
		if expose.IngressDomain == "" {
			return fmt.Errorf("mode is Ingress but ingressDomain is empty")
		}
	} else if expose.Mode == "NodePort" {
		if expose.NodePort == 0 {
			r.logger.Info("mode is NodePort and nodePort is not set, " +
				"nodePort will be a random number between 30000 and 32767")
		}
		if expose.NodePort < 30000 || expose.NodePort > 32767 {
			return fmt.Errorf("invalid NodePort %d, "+
				"must be between 30000–32767", expose.NodePort)
		}
	} else {
		return fmt.Errorf("expose mode %s is not supported", expose.Mode)
	}
	return nil
}
