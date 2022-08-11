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
	// 1.enable_ingress 默认为 false, webhook 将设置相反的值
	// 2.当 设置 enable_ingress 为 true 时, enable_service 必须为 true
	// 将通过 admission webhook 来解决.

	logger := log.FromContext(ctx)

	horus := &horusiov1beta1.Horus{}

	// 从缓存中获取 Horus
	// Get retrieves an obj for the given object key from the Kubernetes Cluster.
	// obj must be a struct pointer so that obj can be updated with the response
	// returned by the Server.
	if err := r.Get(ctx, req.NamespacedName, horus); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// 根据 Horus 的配置进行处理

	// 1.Deployment 的处理
	deployment := utils.NewDeployment(horus)
	// 将 deployment 的 ownerReferences 设置为 horus, 这样, 当删除 horus 时,
	// deployment 也会被自动删除掉.
	if err := controllerutil.SetControllerReference(horus, deployment, r.Scheme); err != nil {
		logger.Error(err, "SetControllerReference for deployment failed")
		return ctrl.Result{}, err
	}
	// 查找同一个 namespace 下的同名的 deployment
	dep := &appsv1.Deployment{}
	if err := r.Get(ctx, req.NamespacedName, dep); err != nil {
		// deployment 不存在则创建
		if errors.IsNotFound(err) {
			if err := r.Create(ctx, deployment); err != nil {
				logger.Error(err, "create deployment failed")
				return ctrl.Result{}, err
			}
		}
		// deployment 存在则更新
	} else {
		if err := r.Update(ctx, deployment); err != nil {
			logger.Error(err, "update deployment failed")
			return ctrl.Result{}, err
		}
	}

	// 2.Service 的处理
	service := utils.NewService(horus)
	if err := controllerutil.SetControllerReference(horus, service, r.Scheme); err != nil {
		logger.Error(err, "SetControllerReference for service failed")
		return ctrl.Result{}, err
	}
	svc := &corev1.Service{}
	if err := r.Get(ctx, types.NamespacedName{Name: horus.Name, Namespace: horus.Namespace}, svc); err != nil {
		if errors.IsNotFound(err) && horus.Spec.EnableService {
			// 这里 Create 的是 service, 而不是 svc.
			if err := r.Create(ctx, service); err != nil {
				logger.Error(err, "create service failed")
				return ctrl.Result{}, err
			}
		}
		// 失败重试
		if !errors.IsNotFound(err) && horus.Spec.EnableService {
			return ctrl.Result{}, err
		}
	} else {
		if horus.Spec.EnableService {
			if err := r.Update(ctx, service); err != nil {
				return ctrl.Result{}, err
			}
		} else {
			// 这里 delete 的不是 service, 而是 svc
			if err := r.Delete(ctx, svc); err != nil {
				logger.Error(err, "delete service failed")
				return ctrl.Result{}, err
			}
		}
	}

	// 3.Ingress 的处理, ingress 配置可能为空
	ingress := utils.NewIngress(horus)
	if err := controllerutil.SetControllerReference(horus, ingress, r.Scheme); err != nil {
		logger.Error(err, "SetControllerReference for ingress failed")
		return ctrl.Result{}, err
	}
	ing := &networkingv1.Ingress{}
	if err := r.Get(ctx, types.NamespacedName{Name: horus.Name, Namespace: horus.Namespace}, ing); err != nil {
		if errors.IsNotFound(err) && horus.Spec.EnableIngress {
			// 这个Create 的是 ingress, 而不是 ing
			if err := r.Create(ctx, ingress); err != nil {
				logger.Error(err, "create ingress failed")
				return ctrl.Result{}, err
			}
		}
		// 失败重试
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
	// 当 Deployment, Service, Ingress 发生变化之后, 就会调用 Reconcile 方法就会被调用
	return ctrl.NewControllerManagedBy(mgr).
		For(&horusiov1beta1.Horus{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&networkingv1.Ingress{}).
		Complete(r)
}
