/*
Copyright 2022 The Arbiter Authors.

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

package executor

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	arbiterv1alpha1 "github.com/kube-arbiter/arbiter/pkg/apis/v1alpha1"
)

// ObservabilityIndicantReconciler reconciles a ObservabilityIndicant object
type ObservabilityIndicantReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=arbiter.k8s.com.cn,resources=observabilityindicants,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=arbiter.k8s.com.cn,resources=observabilityindicants/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=arbiter.k8s.com.cn,resources=observabilityindicants/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// User should Modify the Reconcile function to compare the state specified by
// the ObservabilityIndicant object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *ObservabilityIndicantReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	instance := &arbiterv1alpha1.ObservabilityIndicant{}
	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	policyName := ""
	if instance.ObjectMeta.Annotations != nil {
		policyName = instance.ObjectMeta.Annotations["observability-action-policy"]
	}
	if policyName == "" {
		logger.Info("instance.ObjectMeta.Annotations[\"observability-action-policy\"] has no value")
		return ctrl.Result{}, nil
	}

	oap := &arbiterv1alpha1.ObservabilityActionPolicy{}
	err = r.Client.Get(ctx, types.NamespacedName{
		Namespace: instance.Namespace,
		Name:      policyName,
	}, oap)
	if err != nil {
		logger.Error(err, "r.Client.Get ObservabilityActionPolicy error")
		if errors.IsNotFound(err) {
			logger.Error(err, "ObservabilityActionPolicy IsNotFound")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if oap.Annotations == nil {
		oap.Annotations = make(map[string]string)
	}
	if oap.Annotations["obi-resource-version"] != instance.ObjectMeta.GetResourceVersion() {
		oap.Annotations["obi-resource-version"] = instance.ObjectMeta.GetResourceVersion()
		if err := r.Client.Update(ctx, oap); err != nil {
			logger.Error(err, "update ObservabilityActionPolicy error")
		}
		logger.Info("obi-resource-version updated for observability-action-policy: ", instance.Namespace+"-"+policyName)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ObservabilityIndicantReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&arbiterv1alpha1.ObservabilityIndicant{}).
		Complete(r)
}
