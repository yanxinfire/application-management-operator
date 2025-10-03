package apps

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/yanxinfire/application-management-operator/api/apps/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func NewResourceFromTemplate[T any](templateName string, app *v1alpha1.Application) *T {
	tmpl, err := template.ParseFiles(fmt.Sprintf("templates/%s.yaml", templateName))
	if err != nil {
		return nil
	}
	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, app)
	if err != nil {
		return nil
	}
	obj := new(T)
	err = yaml.Unmarshal(buf.Bytes(), obj)
	if err != nil {
		return nil
	}
	return obj
}

func NewDeployment(app *v1alpha1.Application) *appsv1.Deployment {
	metaData := NewMetadata(app)
	deployment := &appsv1.Deployment{
		ObjectMeta: metaData,
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: metaData.GetLabels(),
			},
			Replicas: app.Spec.Replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metaData,
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            app.Name,
							Image:           app.Spec.Image,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: app.Spec.Port,
									Protocol:      corev1.ProtocolTCP,
								},
							},
						},
					},
				},
			},
		},
	}
	return deployment
}

func NewService(app *v1alpha1.Application) *corev1.Service {
	metaData := NewMetadata(app)
	service := &corev1.Service{
		ObjectMeta: metaData,
		Spec: corev1.ServiceSpec{
			Selector: metaData.GetLabels(),
			Ports:    []corev1.ServicePort{},
			Type:     corev1.ServiceTypeClusterIP,
		},
	}
	port := corev1.ServicePort{
		Port:       app.Spec.Expose.ServicePort,
		TargetPort: intstr.FromString("http"),
		Protocol:   corev1.ProtocolTCP,
	}
	if app.Spec.Expose.Mode == "NodePort" {
		port.NodePort = app.Spec.Expose.NodePort
		service.Spec.Type = corev1.ServiceTypeNodePort
	}
	service.Spec.Ports = append(service.Spec.Ports, port)
	return service
}

func NewIngress(app *v1alpha1.Application) *networkingv1.Ingress {
	metaData := NewMetadata(app)
	metaData.SetAnnotations(map[string]string{
		"kubernetes.io/ingress.class": "nginx",
	})
	ingClass := "nginx"
	ingress := &networkingv1.Ingress{
		ObjectMeta: metaData,
		Spec: networkingv1.IngressSpec{
			IngressClassName: &ingClass,
			Rules:            []networkingv1.IngressRule{},
		},
	}
	ingRule := networkingv1.IngressRule{Host: app.Spec.Expose.IngressDomain}
	ingRule.HTTP = &networkingv1.HTTPIngressRuleValue{
		Paths: []networkingv1.HTTPIngressPath{},
	}
	path1 := networkingv1.HTTPIngressPath{
		Path: "/",
		Backend: networkingv1.IngressBackend{
			Service: &networkingv1.IngressServiceBackend{
				Name: app.Name,
				Port: networkingv1.ServiceBackendPort{
					Number: app.Spec.Expose.ServicePort,
				},
			},
		},
	}
	path1PathType := networkingv1.PathTypePrefix
	path1.PathType = &path1PathType
	ingRule.HTTP.Paths = append(ingRule.HTTP.Paths, path1)
	ingress.Spec.Rules = append(ingress.Spec.Rules, ingRule)
	return ingress
}

func NewMetadata(app *v1alpha1.Application) metav1.ObjectMeta {
	labels := map[string]string{
		"app": app.Name,
	}
	return metav1.ObjectMeta{
		Name:      app.Name,
		Namespace: app.Namespace,
		Labels:    labels,
	}
}
