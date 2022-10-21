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
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/Knetic/govaluate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	arbiterv1alpha1 "github.com/kube-arbiter/arbiter/pkg/apis/v1alpha1"
	pb "github.com/kube-arbiter/arbiter/pkg/proto/lib/executor"
	"github.com/kube-arbiter/arbiter/pkg/utils"
)

const (
	protocol = "unix"
	sockAddr = "/plugins/"

	finalizer = "executor.arbiter.k8s.com.cn"
)

// ObservabilityActionPolicyReconciler reconciles a ObservabilityActionPolicy object
type ObservabilityActionPolicyReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	SockAddr string
}

func fileExist(path string) bool {
	_, err := os.Lstat(path)
	return !os.IsNotExist(err)
}

//+kubebuilder:rbac:groups=arbiter.k8s.com.cn,resources=observabilityactionpolicies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=arbiter.k8s.com.cn,resources=observabilityactionpolicies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=arbiter.k8s.com.cn,resources=observabilityactionpolicies/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// User should modify the Reconcile function to compare the state specified by
// the ObservabilityActionPolicy object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *ObservabilityActionPolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	instance := &arbiterv1alpha1.ObservabilityActionPolicy{}
	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("ObservabilityActionPolicy IsNotFound")
			return ctrl.Result{}, nil
		} else {
			logger.Error(err, "r.Client.Get ObservabilityActionPolicy")
		}
		return ctrl.Result{}, err
	}

	obi := &arbiterv1alpha1.ObservabilityIndicant{}
	err = r.Client.Get(ctx, types.NamespacedName{
		Namespace: instance.Namespace,
		Name:      instance.Spec.ObIndicantName,
	}, obi)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("ObservabilityIndicant IsNotFound", "ObIndicantName", instance.Spec.ObIndicantName)
			return ctrl.Result{}, nil
		} else {
			logger.Error(err, "r.Client.Get ObservabilityIndicant")
		}
		return ctrl.Result{}, err
	}
	//sockFile := sockAddr + instance.Spec.ActionProvider + ".sock"
	sockFile := r.SockAddr

	if instance.DeletionTimestamp.IsZero() {
		if !utils.ContainFinalizers(instance.Finalizers, finalizer) {
			instance.Finalizers = utils.AddFinalizers(instance.Finalizers, finalizer)
			err = r.Client.Update(ctx, instance)
			return reconcile.Result{Requeue: true}, err
		}

		logger.V(4).Info("ObIndicantName", "ObIndicantName", instance.Spec.ObIndicantName)
		if obi.Annotations["observability-action-policy"] != instance.Name {
			obi.Annotations["observability-action-policy"] = instance.Name
			if err := r.Client.Update(ctx, obi); err != nil {
				logger.Error(err, "update ObservabilityIndicant",
					"ObIndicantName", instance.Spec.ObIndicantName)
			}
		}

		functions := map[string]govaluate.ExpressionFunction{
			// NOTE: interface  real type is []string.
			"avg": func(args ...interface{}) (interface{}, error) {
				s, ok := args[0].([]string)
				if !ok {
					logger.Info("avg args[0] type error", "args[0]", args[0])
					return float64(0), nil
				}
				if len(s) == 0 {
					return float64(0), nil
				}

				l := len(s)
				var sum float64 = 0
				for _, sval := range s {
					if sval == "" {
						l--
						continue
					}
					val, err := strconv.ParseFloat(sval, 64)
					if err != nil {
						logger.Error(err, "evaluate error")
						return ctrl.Result{}, err
					}
					sum += val
				}
				if l == 0 {
					return float64(0), nil
				}
				logger.V(4).Info("sum=", sum)
				return sum / (float64(l)), nil
			},
		}

		actionInfo := []arbiterv1alpha1.ActionInfo{}

		// 这个map存储的就是cpu, mem的时间序列数字， 都是[]string。
		parameters := make(map[string]interface{})
		// 目前新定义的CR，在status中不会存储pod名字。名字只能通过targetRef来获取。所以支起拿的循环status逻辑不再适合
		// 但是，需要注意，新的CRD的定义中，加入了labels。这个目前虽然有字段，但是不支持
		for metricName, poi := range obi.Status.Metrics {
			// NOTE: 如果使用了query查询，只需要获取到返回的数据的列表
			// 或者说，不管之前的数据做了何种聚合，我这里只是对获取到的数据做一个加权
			path := fmt.Sprintf("metrics.%s", metricName)
			calculateNums := make([]string, 0)
			for _, values := range poi {
				for _, value := range values.Records {
					calculateNums = append(calculateNums, value.Value)
				}
			}
			parameters[path] = calculateNums
		}
		expr, err := govaluate.NewEvaluableExpressionWithFunctions(instance.Spec.Condition.Expression, functions)
		if err != nil {
			logger.Error(err, "govaluate.NewEvaluableExpressionWithFunctions")
			return ctrl.Result{}, err
		}

		expressionValue, err := expr.Evaluate(parameters)
		if err != nil {
			logger.Error(err, "expr.Evaluate(parameters)")
			return ctrl.Result{}, err
		}

		expr, err = govaluate.NewEvaluableExpressionWithFunctions(instance.Spec.Condition.Expression+" "+instance.Spec.Condition.Operator+" "+instance.Spec.Condition.TargetValue, functions)
		if err != nil {
			logger.Error(err, "govaluate.NewEvaluableExpressionWithFunctions")
			return ctrl.Result{}, err
		}

		conditionValue, err := expr.Evaluate(parameters)
		if err != nil {
			logger.Error(err, "expr.Evaluate(parameters)")
			return ctrl.Result{}, err
		}

		pai := arbiterv1alpha1.ActionInfo{}

		// Maybe pod or node,
		resourceName, err := r.FindResourceNameByIndex(ctx, obi)
		if resourceName == "" {
			return ctrl.Result{}, err
		}
		pai.ResourceName = resourceName
		pai.ExpressionValue = strconv.FormatFloat(expressionValue.(float64), 'f', 2, 64)
		pai.ConditionValue = strconv.FormatBool(conditionValue.(bool))

		actionInfo = append(actionInfo, pai)

		logger.V(4).Info("instance.Spec.Condition.Expression", "Expression", expressionValue)

		needUpdate := false
		if len(actionInfo) != len(instance.Status.ActionInfo) {
			needUpdate = true
		} else {
			for i := 0; i < len(actionInfo); i++ {
				if actionInfo[i].ResourceName != instance.Status.ActionInfo[i].ResourceName ||
					actionInfo[i].ExpressionValue != instance.Status.ActionInfo[i].ExpressionValue ||
					actionInfo[i].ConditionValue != instance.Status.ActionInfo[i].ConditionValue {
					needUpdate = true
					break
				}
			}
		}

		if needUpdate {
			instance.Status.ActionInfo = actionInfo
			err = r.Client.Update(ctx, instance)
			if err != nil {
				return ctrl.Result{}, err
			}
		}

		// Call Action Provider plugin
		count := 1
		for {
			if fileExist(sockFile) {
				break
			}

			logger.Error(err, "sock file not exist yet, retry",
				"sockFile", sockFile, "retry", count)
			time.Sleep(5 * time.Second)
			count++
		}

		conn, err := ExecuteConn(sockFile)
		if err != nil {
			logger.Error(err, "try to connect server")
			return reconcile.Result{}, err
		}
		defer conn.Close()
		var (
			client  = pb.NewExecuteClient(conn)
			rootCtx = ctx
			timeout = 30 * time.Second
		)

		for i := 0; i < len(actionInfo); i++ {
			func() {
				ctx, cancel := context.WithTimeout(rootCtx, timeout)
				defer cancel()

				exprVal, err := strconv.ParseFloat(actionInfo[i].ExpressionValue, 64)
				if err != nil {
					logger.Error(err, "strconv.ParseFloat")
					exprVal = 0
				}

				condVal, err := strconv.ParseBool(actionInfo[i].ConditionValue)
				if err != nil {
					logger.Error(err, "strconv.ParseBool")
					exprVal = 0
				}
				message := pb.ExecuteMessage{ResourceName: actionInfo[i].ResourceName,
					Namespace:  obi.Spec.TargetRef.Namespace,
					ExprVal:    exprVal,
					CondVal:    condVal,
					ActionData: instance.Spec.ActionData,
					Group:      obi.Spec.TargetRef.Group,
					Version:    obi.Spec.TargetRef.Version,
					Resources:  "pods",
					Kind:       pb.Kind_pod,
				}
				if obi.Spec.TargetRef.Kind == "Node" {
					message.Resources = "nodes"
					message.Kind = pb.Kind_node
				}

				res, err := client.Execute(ctx, &message)
				if err != nil {
					logger.Error(err, "client.Execute")
				}
				logger.V(4).Info("Got result from executor plugin", "res", res)
			}()
		}

		return ctrl.Result{}, nil
	}

	logger.Info("Delete Policy. Which will remove finalzer and delete labels",
		"policy", instance.GetName())
	instance.Finalizers = utils.RemoveFinalizer(instance.Finalizers, finalizer)
	resourceName, err := r.FindResourceNameByIndex(ctx, obi)
	if resourceName == "" {
		return ctrl.Result{}, err
	}
	message := pb.ExecuteMessage{
		ResourceName: resourceName,
		CondVal:      false,
		ActionData:   instance.Spec.ActionData,
		Namespace:    obi.Spec.TargetRef.Namespace,
		Group:        obi.Spec.TargetRef.Group,
		Version:      obi.Spec.TargetRef.Version,
		Resources:    "pods", // TODO: now only support nodes and pods, will support more in v0.2.0
		Kind:         pb.Kind_pod,
	}
	if obi.Spec.TargetRef.Kind == "Node" {
		message.Resources = "nodes"
		message.Kind = pb.Kind_node
	}

	conn, err := ExecuteConn(sockFile)
	if err != nil {
		logger.Error(err, "try to connect server")
		return reconcile.Result{}, err
	}
	defer conn.Close()
	client := pb.NewExecuteClient(conn)
	_, err = client.Execute(ctx, &message)
	if err != nil {
		logger.Error(err, "client.Execute")
		return reconcile.Result{}, err
	}

	err = r.Client.Update(ctx, instance)
	return reconcile.Result{Requeue: true}, err
}

func ExecuteConn(sockeFilePath string) (*grpc.ClientConn, error) {
	dialer := func(c context.Context, addr string) (net.Conn, error) {
		return net.Dial(protocol, addr)
	}

	return grpc.Dial(sockeFilePath, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(dialer))
}

// SetupWithManager sets up the controller with the Manager.
func (r *ObservabilityActionPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&arbiterv1alpha1.ObservabilityActionPolicy{}).
		Complete(r)
}

func (r *ObservabilityActionPolicyReconciler) FindResourceNameByIndex(ctx context.Context, obi *arbiterv1alpha1.ObservabilityIndicant) (string, error) {
	resourceName := obi.Spec.TargetRef.Name
	if resourceName != "" {
		return resourceName, nil
	}

	u := unstructured.UnstructuredList{Object: map[string]interface{}{
		"apiVersion": "v1",
		"kind":       obi.Spec.TargetRef.Kind,
	}}
	options := []client.ListOption{client.MatchingLabels(obi.Spec.TargetRef.Labels)}
	if kind := obi.Spec.TargetRef.Kind; kind == "Pod" {
		options = append(options, client.InNamespace(obi.Spec.TargetRef.Namespace))
	}

	if err := r.Client.List(ctx, &u, options...); err != nil {
		if errors.IsNotFound(err) {
			klog.Errorf("can't find resource by %s/%s/%s\n",
				obi.Spec.TargetRef.Group, obi.Spec.TargetRef.Version, obi.Spec.TargetRef.Kind)
			return "", nil
		}
		return "", err
	}
	if uint64(len(u.Items)) <= obi.Spec.TargetRef.Index {
		klog.Errorf("index out of range [%d] with length %d\n", obi.Spec.TargetRef.Index, len(u.Items))
		return "", nil
	}

	return u.Items[obi.Spec.TargetRef.Index].GetName(), nil
}
