package utils

import (
	"bytes"
	"path/filepath"
	"text/template"

	horusiov1beta1 "github.com/horus/api/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// NewDeployment constructs a deployment resource for horus resource.
func NewDeployment(horus *horusiov1beta1.Horus) *appsv1.Deployment {
	deploy := &appsv1.Deployment{}
	yaml.Unmarshal(parseTemplate("deployment", horus), deploy)
	return deploy
}

// NewService constructs a service resource for hours resource.
func NewService(horus *horusiov1beta1.Horus) *corev1.Service {
	svc := &corev1.Service{}
	yaml.Unmarshal(parseTemplate("service", horus), svc)
	return svc
}

// NewIngress constructs a Ingress resource for horus resource.
func NewIngress(horus *horusiov1beta1.Horus) *networkingv1.Ingress {
	ing := &networkingv1.Ingress{}
	yaml.Unmarshal(parseTemplate("ingress", horus), ing)
	return ing
}

// parseTemplate will parse the templates definitions from "controller/templates/xxx.yaml"
// and applies parsed templates to the deploy/service/ingress resource.
func parseTemplate(kind string, horus *horusiov1beta1.Horus) []byte {
	tpl, err := template.ParseFiles(filepath.Join("controllers/template/", kind+".yaml"))
	if err != nil {
		panic(err)
	}

	buf := new(bytes.Buffer)
	err = tpl.Execute(buf, horus)
	if err != nil {
		panic(err)
	}

	return buf.Bytes()
}
