package apps

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/yanxinfire/application-management-operator/api/apps/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func parseTemplate(templateName string, app v1alpha1.Application) []byte {
	tmpl, err := template.ParseFiles(fmt.Sprintf("internal/controller/templates/%s.yaml", templateName))
	if err != nil {
		return []byte{}
	}
	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, app)
	if err != nil {
		return []byte{}
	}
	return buf.Bytes()
}

func NewDeployment(app v1alpha1.Application) (*appsv1.Deployment, error) {
	deployment := &appsv1.Deployment{}
	err := yaml.Unmarshal(parseTemplate("deployment", app), deployment)
	if err != nil {
		return nil, err
	}
	return deployment, nil
}

func NewService(app v1alpha1.Application) (*corev1.Service, error) {
	svc := &corev1.Service{}
	err := yaml.Unmarshal(parseTemplate("service", app), svc)
	if err != nil {
		return nil, err
	}
	return svc, nil
}

func NewIngress(app v1alpha1.Application) (*networkingv1.Ingress, error) {
	ing := &networkingv1.Ingress{}
	err := yaml.Unmarshal(parseTemplate("ingress", app), ing)
	if err != nil {
		return nil, err
	}
	return ing, nil
}
