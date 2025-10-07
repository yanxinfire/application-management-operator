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
	"reflect"
	"time"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	appsv1alpha1 "github.com/yanxinfire/application-management-operator/api/apps/v1alpha1"
)

const (
	STATUSREADY    = "ready"
	STATUSNOTREADY = "not ready"
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
	appCopy := app.DeepCopy()
	appCopy.Status = appsv1alpha1.ApplicationStatus{
		Deployment: STATUSNOTREADY,
		Service:    STATUSNOTREADY,
		Ingress:    STATUSNOTREADY,
	}

	if err := r.VerifyApplicationMode(appCopy); err != nil {
		appCopy.Status.Reason = err.Error()
		_ = r.UpdateStatus(ctx, appCopy, app)
		return ctrl.Result{}, err
	}

	if err := r.CreateOrUpdateDeployment(ctx, appCopy); err != nil {
		appCopy.Status.Reason = err.Error()
		_ = r.UpdateStatus(ctx, appCopy, app)
		return ctrl.Result{RequeueAfter: 30 * time.Second}, err
	}

	if err := r.CreateOrUpdateService(ctx, appCopy); err != nil {
		appCopy.Status.Reason = err.Error()
		_ = r.UpdateStatus(ctx, appCopy, app)
		return ctrl.Result{RequeueAfter: 30 * time.Second}, err
	}
	appCopy.Status.Service = STATUSNOTREADY
	if app.Spec.Expose.Mode == "Ingress" {
		if err := r.createOrUpdateIngress(ctx, appCopy); err != nil {
			appCopy.Status.Reason = err.Error()
			_ = r.UpdateStatus(ctx, appCopy, app)
			return ctrl.Result{RequeueAfter: 30 * time.Second}, err
		}
		appCopy.Status.Ingress = STATUSNOTREADY
	} else {
		if err := r.DeleteIngress(ctx, appCopy); err != nil {
			appCopy.Status.Reason = err.Error()
			_ = r.UpdateStatus(ctx, appCopy, app)
			return ctrl.Result{RequeueAfter: 30 * time.Second}, err
		}
	}

	err := r.UpdateStatus(ctx, appCopy, app)
	if err != nil {
		return ctrl.Result{RequeueAfter: 10}, err
	}

	return ctrl.Result{}, nil
}

func (r *ApplicationReconciler) CreateOrUpdateDeployment(
	ctx context.Context, app *appsv1alpha1.Application) error {
	deployment := NewDeployment(app)
	err := controllerutil.SetControllerReference(app, deployment, r.Scheme)
	if err != nil {
		return err
	}
	existingDeployment := &appsv1.Deployment{}
	if err = r.Get(ctx, types.NamespacedName{
		Namespace: app.Namespace,
		Name:      app.Name,
	}, existingDeployment); err != nil {
		if errors.IsNotFound(err) {
			r.logger.Info("Creating Deployment", "Namespace",
				app.Namespace, "Name", app.Name)
			return r.Create(ctx, deployment)
		}
		return err
	}

	// Utilise --dry-run='client' to update deployment unsetting properties,
	// so that it could be compared with existing deployment correctly
	err = r.Update(ctx, deployment, client.DryRunAll)
	if err != nil {
		return err
	}
	if !equality.Semantic.DeepEqual(deployment.Spec, existingDeployment.Spec) {
		r.logger.Info("Updating Deployment", "Namespace",
			app.Namespace, "Name", app.Name)
		return r.Update(ctx, deployment)
	}
	if existingDeployment.Status.ReadyReplicas == app.Spec.Replicas {
		app.Status.Deployment = STATUSNOTREADY
	}
	return nil
}

func (r *ApplicationReconciler) CreateOrUpdateService(
	ctx context.Context, app *appsv1alpha1.Application) error {
	service := NewService(app)
	err := controllerutil.SetControllerReference(app, service, r.Scheme)
	if err != nil {
		return err
	}
	existingService := &corev1.Service{}
	if err = r.Get(ctx, types.NamespacedName{
		Namespace: app.Namespace,
		Name:      app.Name,
	}, existingService); err != nil {
		if errors.IsNotFound(err) {
			r.logger.Info("Creating Service", "Namespace",
				app.Namespace, "Name", app.Name)
			return r.Create(ctx, service, client.FieldOwner(app.Name))
		}
		return err
	}

	err = r.Update(ctx, service, client.DryRunAll)
	if err != nil {
		return err
	}
	if !equality.Semantic.DeepEqual(service.Spec, existingService.Spec) {
		r.logger.Info("Updating Service", "Namespace",
			app.Namespace, "Name", app.Name)
		return r.Update(ctx, service, client.FieldOwner(app.Name))
	}
	return nil
}

func (r *ApplicationReconciler) createOrUpdateIngress(
	ctx context.Context, app *appsv1alpha1.Application) error {
	ingress := NewIngress(app)
	err := controllerutil.SetControllerReference(app, ingress, r.Scheme)
	if err != nil {
		return err
	}

	existingIngress := &networkingv1.Ingress{}
	if err = r.Get(ctx, types.NamespacedName{
		Namespace: app.Namespace,
		Name:      app.Name,
	}, existingIngress); err != nil {
		if errors.IsNotFound(err) {
			r.logger.Info("Creating Ingress", "Namespace",
				app.Namespace, "Name", app.Name)
			return r.Create(ctx, ingress)
		}
		return err
	}

	err = r.Update(ctx, ingress, client.DryRunAll)
	if err != nil {
		return err
	}
	if !equality.Semantic.DeepEqual(ingress.Spec, existingIngress.Spec) {
		r.logger.Info("Updating Ingress", "Namespace",
			app.Namespace, "Name", app.Name)
		return r.Update(ctx, ingress)
	}
	return nil
}

func (r *ApplicationReconciler) DeleteIngress(ctx context.Context, app *appsv1alpha1.Application) error {
	ingress := &networkingv1.Ingress{}
	if err := r.Get(ctx, types.NamespacedName{
		Namespace: app.Namespace,
		Name:      app.Name,
	}, ingress); err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	r.logger.Info("Deleting Ingress", "Namespace",
		app.Namespace, "Name", app.Name)
	return r.Delete(ctx, ingress)
}

func (r *ApplicationReconciler) UpdateStatus(ctx context.Context, appCopy, app *appsv1alpha1.Application) error {
	if !reflect.DeepEqual(appCopy.Status, app.Status) {
		// return r.Status().Update(ctx, appCopy)
		return r.Update(ctx, appCopy)
	}
	return nil
}

func (r *ApplicationReconciler) VerifyApplicationMode(app *appsv1alpha1.Application) error {
	expose := app.Spec.Expose
	switch expose.Mode {
	case "Ingress":
		if expose.IngressDomain == "" {
			return fmt.Errorf("mode is Ingress but ingressDomain is empty")
		}
		return nil
	case "NodePort":
		if expose.NodePort == 0 {
			r.logger.Info("mode is NodePort and nodePort is not set, " +
				"nodePort will be a random number between 30000 and 32767")
		}
		if expose.NodePort < 30000 || expose.NodePort > 32767 {
			return fmt.Errorf("invalid NodePort %d, "+
				"must be between 30000–32767", expose.NodePort)
		}
		return nil
	}
	return fmt.Errorf("expose mode %s is not supported", expose.Mode)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.Application{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&networkingv1.Ingress{}).
		Named("apps-application").
		Complete(r)
}
