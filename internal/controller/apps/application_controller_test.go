package apps

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1alpha1 "github.com/yanxinfire/application-management-operator/api/apps/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// --- VerifyApplicationMode() ---
func TestVerifyApplicationMode(t *testing.T) {
	r := &ApplicationReconciler{}

	tests := []struct {
		name    string
		app     *appsv1alpha1.Application
		wantErr bool
	}{
		{
			name: "valid Ingress",
			app: &appsv1alpha1.Application{
				Spec: appsv1alpha1.ApplicationSpec{
					Expose: &appsv1alpha1.Expose{
						Mode:          "Ingress",
						IngressDomain: "example.com",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Ingress missing domain",
			app: &appsv1alpha1.Application{
				Spec: appsv1alpha1.ApplicationSpec{
					Expose: &appsv1alpha1.Expose{Mode: "Ingress"},
				},
			},
			wantErr: true,
		},
		{
			name: "NodePort valid",
			app: &appsv1alpha1.Application{
				Spec: appsv1alpha1.ApplicationSpec{
					Expose: &appsv1alpha1.Expose{Mode: "NodePort", NodePort: 30001},
				},
			},
			wantErr: false,
		},
		{
			name: "NodePort invalid range",
			app: &appsv1alpha1.Application{
				Spec: appsv1alpha1.ApplicationSpec{
					Expose: &appsv1alpha1.Expose{Mode: "NodePort", NodePort: 29999},
				},
			},
			wantErr: true,
		},
		{
			name: "unsupported mode",
			app: &appsv1alpha1.Application{
				Spec: appsv1alpha1.ApplicationSpec{
					Expose: &appsv1alpha1.Expose{Mode: "InvalidMode"},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := r.VerifyApplicationMode(tt.app)
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifyApplicationMode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// --- UpdateStatus() ---
func TestUpdateStatus(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = appsv1alpha1.AddToScheme(scheme)

	app := &appsv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: "app1", Namespace: "default"},
		Status:     appsv1alpha1.ApplicationStatus{Deployment: "ready"},
	}

	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(app).Build()

	appCopy := app.DeepCopy()
	appCopy.Status.Deployment = "not ready"
	r := &ApplicationReconciler{Client: client, Scheme: scheme}

	if err := r.UpdateStatus(context.Background(), appCopy, app); err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}

	got := &appsv1alpha1.Application{}
	_ = client.Get(context.Background(), types.NamespacedName{Name: "app1", Namespace: "default"}, got)
	t.Log(got)

	if got.Status.Deployment != "not ready" {
		t.Errorf("expected status 'not ready', got %s", got.Status.Deployment)
	}
}

// --- CreateOrUpdateDeployment() basic creation ---
func TestCreateOrUpdateDeployment_CreatesWhenMissing(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = appsv1alpha1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)

	app := &appsv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: "demo", Namespace: "default"},
		Spec: appsv1alpha1.ApplicationSpec{
			Image:    "nginx",
			Replicas: 1,
			Port:     8080,
			Env: []corev1.EnvVar{
				{
					Name:  "foo",
					Value: "bar",
				},
			},
		},
	}
	client := fake.NewClientBuilder().WithScheme(scheme).Build()

	r := &ApplicationReconciler{Client: client, Scheme: scheme}

	err := r.CreateOrUpdateDeployment(context.Background(), app)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	deploy := &appsv1.Deployment{}
	err = client.Get(context.Background(), types.NamespacedName{Name: "demo", Namespace: "default"}, deploy)
	if err != nil {
		t.Errorf("deployment not created: %v", err)
	}
	assert.Equal(t, "nginx", deploy.Spec.Template.Spec.Containers[0].Image)
	assert.Equal(t, []corev1.EnvVar{{Name: "foo", Value: "bar"}}, deploy.Spec.Template.Spec.Containers[0].Env)
	assert.Equal(t, int32(8080), deploy.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort)
	assert.Equal(t, int32(1), *deploy.Spec.Replicas)
}

// --- DeleteIngress() ---
func TestDeleteIngress(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = networkingv1.AddToScheme(scheme)

	app := &appsv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: "app2", Namespace: "default"},
	}
	ing := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{Name: "app2", Namespace: "default"},
	}
	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(ing).Build()
	r := &ApplicationReconciler{Client: client, Scheme: scheme}

	if err := r.DeleteIngress(context.Background(), app); err != nil {
		t.Fatalf("DeleteIngress error: %v", err)
	}

	err := client.Get(context.Background(), types.NamespacedName{Name: "app2", Namespace: "default"}, ing)
	if err == nil {
		t.Errorf("ingress should be deleted, but found")
	}
}

// --- Reconcile() sanity check ---
func TestReconcileHappyPath(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = networkingv1.AddToScheme(scheme)
	_ = appsv1alpha1.AddToScheme(scheme)

	app := &appsv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: "myapp", Namespace: "default"},
		Spec: appsv1alpha1.ApplicationSpec{
			Expose:   &appsv1alpha1.Expose{Mode: "NodePort", NodePort: 30001},
			Replicas: 1,
		},
	}

	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(app).Build()
	r := &ApplicationReconciler{Client: client, Scheme: scheme}

	_, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: "myapp", Namespace: "default"},
	})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}
}
