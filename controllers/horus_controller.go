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

package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	horusiov1beta1 "github.com/horus/api/v1beta1"
	"github.com/horus/controllers/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
)

// HorusReconciler reconciles a Horus object
type HorusReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=horus.io,resources=horus,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=horus.io,resources=horus/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=horus.io,resources=horus/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments;statefulsets;daemonsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Horus object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *HorusReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// 1.enable_ingress ????????? false, webhook ?????????????????????
	// 2.??? ?????? enable_ingress ??? true ???, enable_service ????????? true
	// ????????? admission webhook ?????????.

	logger := log.FromContext(ctx)

	horus := &horusiov1beta1.Horus{}

	// ?????????????????? Horus
	// Get retrieves an obj for the given object key from the Kubernetes Cluster.
	// obj must be a struct pointer so that obj can be updated with the response
	// returned by the Server.
	if err := r.Get(ctx, req.NamespacedName, horus); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// ?????? Horus ?????????????????????

	// 1.Deployment ?????????
	deployment := utils.NewDeployment(horus)
	// ??? deployment ??? ownerReferences ????????? horus, ??????, ????????? horus ???,
	// deployment ????????????????????????.
	if err := controllerutil.SetControllerReference(horus, deployment, r.Scheme); err != nil {
		logger.Error(err, "SetControllerReference for deployment failed")
		return ctrl.Result{}, err
	}
	// ??????????????? namespace ??????????????? deployment
	dep := &appsv1.Deployment{}
	if err := r.Get(ctx, req.NamespacedName, dep); err != nil {
		// deployment ??????????????????
		if errors.IsNotFound(err) {
			if err := r.Create(ctx, deployment); err != nil {
				logger.Error(err, "create deployment failed")
				return ctrl.Result{}, err
			}
		}
		// deployment ???????????????
	} else {
		if err := r.Update(ctx, deployment); err != nil {
			logger.Error(err, "update deployment failed")
			return ctrl.Result{}, err
		}
	}

	// 2.Service ?????????
	service := utils.NewService(horus)
	if err := controllerutil.SetControllerReference(horus, service, r.Scheme); err != nil {
		logger.Error(err, "SetControllerReference for service failed")
		return ctrl.Result{}, err
	}
	svc := &corev1.Service{}
	if err := r.Get(ctx, types.NamespacedName{Name: horus.Name, Namespace: horus.Namespace}, svc); err != nil {
		if errors.IsNotFound(err) && horus.Spec.EnableService {
			// ?????? Create ?????? service, ????????? svc.
			if err := r.Create(ctx, service); err != nil {
				logger.Error(err, "create service failed")
				return ctrl.Result{}, err
			}
		}
		// ????????????
		if !errors.IsNotFound(err) && horus.Spec.EnableService {
			return ctrl.Result{}, err
		}
	} else {
		if horus.Spec.EnableService {
			if err := r.Update(ctx, service); err != nil {
				return ctrl.Result{}, err
			}
		} else {
			// ?????? delete ????????? service, ?????? svc
			if err := r.Delete(ctx, svc); err != nil {
				logger.Error(err, "delete service failed")
				return ctrl.Result{}, err
			}
		}
	}

	// 3.Ingress ?????????, ingress ??????????????????
	ingress := utils.NewIngress(horus)
	if err := controllerutil.SetControllerReference(horus, ingress, r.Scheme); err != nil {
		logger.Error(err, "SetControllerReference for ingress failed")
		return ctrl.Result{}, err
	}
	ing := &networkingv1.Ingress{}
	if err := r.Get(ctx, types.NamespacedName{Name: horus.Name, Namespace: horus.Namespace}, ing); err != nil {
		if errors.IsNotFound(err) && horus.Spec.EnableIngress {
			// ??????Create ?????? ingress, ????????? ing
			if err := r.Create(ctx, ingress); err != nil {
				logger.Error(err, "create ingress failed")
				return ctrl.Result{}, err
			}
		}
		// ????????????
		if !errors.IsNotFound(err) && horus.Spec.EnableIngress {
			return ctrl.Result{}, err
		}
	} else {
		if horus.Spec.EnableIngress {
			if err := r.Update(ctx, service); err != nil {
				return ctrl.Result{}, err
			}
		} else {
			if err := r.Delete(ctx, ing); err != nil {
				logger.Error(err, "delete ingress failed")
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HorusReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// ??? Deployment, Service, Ingress ??????????????????, ???????????? Reconcile ?????????????????????
	return ctrl.NewControllerManagedBy(mgr).
		For(&horusiov1beta1.Horus{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&networkingv1.Ingress{}).
		Complete(r)
}
