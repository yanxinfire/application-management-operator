package apps

import (
	"reflect"
	"testing"

	"github.com/yanxinfire/application-management-operator/api/apps/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
)

func TestNewDeployment(t *testing.T) {
	type args struct {
		app v1alpha1.Application
	}
	tests := []struct {
		name    string
		args    args
		want    *appsv1.Deployment
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewDeployment(tt.args.app)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDeployment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDeployment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewService(t *testing.T) {
	type args struct {
		app v1alpha1.Application
	}
	tests := []struct {
		name    string
		args    args
		want    *corev1.Service
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewService(tt.args.app)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewService() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewService() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewIngress(t *testing.T) {
	type args struct {
		app v1alpha1.Application
	}
	tests := []struct {
		name    string
		args    args
		want    *networkingv1.Ingress
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewIngress(tt.args.app)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewIngress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewIngress() = %v, want %v", got, tt.want)
			}
		})
	}
}
