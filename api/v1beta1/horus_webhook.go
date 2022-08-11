/*
Copyright 2022.

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

package v1beta1

import (
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var horuslog = logf.Log.WithName("horus-resource")

func (r *Horus) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-horus-io-v1beta1-horus,mutating=true,failurePolicy=fail,sideEffects=None,groups=horus.io,resources=horus,verbs=create;update,versions=v1beta1,name=mhorus.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Horus{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Horus) Default() {
	horuslog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
	// 把 ingress 的值设置成反向的
	r.Spec.EnableIngress = !r.Spec.EnableIngress
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-horus-io-v1beta1-horus,mutating=false,failurePolicy=fail,sideEffects=None,groups=horus.io,resources=horus,verbs=create;update,versions=v1beta1,name=vhorus.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Horus{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Horus) ValidateCreate() error {
	horuslog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return r.validateHorus()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Horus) ValidateUpdate(old runtime.Object) error {
	horuslog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return r.validateHorus()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Horus) ValidateDelete() error {
	horuslog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	// 不做任何处理
	return nil
}

func (r *Horus) validateHorus() error {
	// 如果 enable_ingress 但是没有 enable_service, 返回错误
	if !r.Spec.EnableService && r.Spec.EnableIngress {
		k8serr.NewInvalid(GroupVersion.WithKind("Horus").GroupKind(), r.Name,
			field.ErrorList{
				field.Invalid(field.NewPath("enable_service"),
					r.Spec.EnableService,
					"enable_service should be true when enable_ingress is true"),
			},
		)
	}
	return nil
}
