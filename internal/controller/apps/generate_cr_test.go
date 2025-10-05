package apps

import (
	"os"
	"reflect"
	"testing"

	"github.com/yanxinfire/application-management-operator/api/apps/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func newResource[T any](filename string) *T {
	b, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	var obj T
	if err := yaml.Unmarshal(b, &obj); err != nil {
		panic(err)
	}
	return &obj
}

func TestNewResourceFromTemplate(t *testing.T) {
	type args struct {
		templateName string
		app          *v1alpha1.Application
	}
	tests := []struct {
		name string
		args args
		want *appsv1.Deployment
	}{
		{
			name: "Test Ingress Application",
			args: args{
				app: newResource[v1alpha1.Application](
					"testdata/app_ing_cr.yaml"),
				templateName: "deployment",
			},
			want: newResource[appsv1.Deployment](
				"testdata/deploy_ing_expect.yaml"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewResourceFromTemplate[appsv1.Deployment](
				tt.args.templateName, tt.args.app); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewDeployment(t *testing.T) {
	type args struct {
		app *v1alpha1.Application
	}
	tests := []struct {
		name string
		args args
		want *appsv1.Deployment
	}{
		{
			name: "Test Deployment Generation",
			args: args{
				newResource[v1alpha1.Application](
					"testdata/app_ing_cr.yaml")},
			want: newResource[appsv1.Deployment](
				"testdata/deploy_ing_expect.yaml"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewDeployment(tt.args.app)
			if !equality.Semantic.DeepEqual(got.Spec, tt.want.Spec) {
				t.Errorf("got %v, want %v",
					got.Spec, tt.want.Spec)
			}
			if !equality.Semantic.DeepEqual(got.ObjectMeta, tt.want.ObjectMeta) {
				t.Errorf("want %v, want %v",
					got.ObjectMeta, tt.want.ObjectMeta)
			}
		})
	}
}

func TestNewService(t *testing.T) {
	type args struct {
		app *v1alpha1.Application
	}
	tests := []struct {
		name string
		args args
		want *corev1.Service
	}{
		{
			name: "Test Ingress Service Generation",
			args: args{
				app: newResource[v1alpha1.Application](
					"testdata/app_ing_cr.yaml"),
			},
			want: newResource[corev1.Service](
				"testdata/svc_ing_expect.yaml"),
		},
		{
			name: "Test NodePort Service Generation",
			args: args{
				app: newResource[v1alpha1.Application](
					"testdata/app_np_cr.yaml"),
			},
			want: newResource[corev1.Service](
				"testdata/svc_np_expect.yaml"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewService(tt.args.app)
			if !equality.Semantic.DeepEqual(got.Spec, tt.want.Spec) {
				t.Errorf("got %v, want %v",
					got.Spec, tt.want.Spec)
			}
			if !equality.Semantic.DeepEqual(got.ObjectMeta, tt.want.ObjectMeta) {
				t.Errorf("got %v, want %v",
					got.ObjectMeta, tt.want.ObjectMeta)
			}
		})
	}
}

func TestNewIngress(t *testing.T) {
	type args struct {
		app *v1alpha1.Application
	}
	tests := []struct {
		name string
		args args
		want *networkingv1.Ingress
	}{
		{
			name: "Test Ingress Generation",
			args: args{
				app: newResource[v1alpha1.Application](
					"testdata/app_ing_cr.yaml"),
			},
			want: newResource[networkingv1.Ingress](
				"testdata/ing_expect.yaml"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewIngress(tt.args.app)
			if !equality.Semantic.DeepEqual(got.Spec, tt.want.Spec) {
				t.Errorf("got %v, want %v",
					got.Spec, tt.want.Spec)
			}
			if !equality.Semantic.DeepEqual(got.ObjectMeta, tt.want.ObjectMeta) {
				t.Errorf("got %v, want %v",
					got.ObjectMeta, tt.want.ObjectMeta)
			}
		})
	}
}

func TestNewMetadata(t *testing.T) {
	type args struct {
		app *v1alpha1.Application
	}
	tests := []struct {
		name string
		args args
		want metav1.ObjectMeta
	}{
		{
			name: "Test Metadata Generation",
			args: args{
				app: newResource[v1alpha1.Application](
					"testdata/app_ing_cr.yaml"),
			},
			want: metav1.ObjectMeta{
				Name:      "my-test-ing",
				Namespace: "my-test",
				Labels: map[string]string{
					"app": "my-test-ing",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewMetadata(tt.args.app); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
