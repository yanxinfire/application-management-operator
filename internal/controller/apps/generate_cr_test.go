package apps

import (
	"os"
	"reflect"
	"testing"

	"github.com/yanxinfire/application-management-operator/api/apps/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
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
		{
			name: "Test NodePort Application",
			args: args{
				templateName: "deployment",
				app: newResource[v1alpha1.Application](
					"testdata/app_np_cr.yaml"),
			},
			want: newResource[appsv1.Deployment](
				"testdata/deploy_np_expect.yaml"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewResourceFromTemplate[appsv1.Deployment](
				tt.args.templateName, tt.args.app); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewResourceFromTemplate() = %v, want %v", got, tt.want)
			}
		})
	}
}
