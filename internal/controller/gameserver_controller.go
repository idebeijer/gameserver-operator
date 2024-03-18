package controller

import (
	"context"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	gameserverv1alpha1 "github.com/idebeijer/gameserver-operator/api/v1alpha1"
)

const (
	typeAvailableGameServer = "Available"
	typeDegradedMemcached   = "Degraded"
)

// GameServerReconciler reconciles a GameServer object
type GameServerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=gameserver.idebeijer.github.io,resources=gameservers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=gameserver.idebeijer.github.io,resources=gameservers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=gameserver.idebeijer.github.io,resources=gameservers/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=events,verbs=create;patch
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *GameServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	gameserver := &gameserverv1alpha1.GameServer{}
	err := r.Get(ctx, req.NamespacedName, gameserver)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	if gameserver.Status.Conditions == nil || len(gameserver.Status.Conditions) == 0 {
		meta.SetStatusCondition(&gameserver.Status.Conditions, metav1.Condition{Type: typeAvailableGameServer, Status: metav1.ConditionUnknown, Reason: "Reconciling", Message: "Reconciling GameServer"})
		if err = r.Status().Update(ctx, gameserver); err != nil {
			return ctrl.Result{}, err
		}

		if err := r.Get(ctx, req.NamespacedName, gameserver); err != nil {
			return ctrl.Result{}, err
		}
	}

	found := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: gameserver.Name, Namespace: gameserver.Namespace}, found)
	if err != nil && apierrors.IsNotFound(err) {
		dep, err := r.deploymentForGameServer(gameserver)
		if err != nil {
			return ctrl.Result{}, err
		}

		if err := r.Create(ctx, dep); err != nil {
			return ctrl.Result{}, err
		}

		meta.SetStatusCondition(&gameserver.Status.Conditions, metav1.Condition{Type: typeAvailableGameServer, Status: metav1.ConditionTrue, Reason: "DeploymentCreated", Message: "Deployment created"})
		if err = r.Status().Update(ctx, gameserver); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{RequeueAfter: time.Minute}, nil
	} else if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *GameServerReconciler) deploymentForGameServer(gs *gameserverv1alpha1.GameServer) (*appsv1.Deployment, error) {
	ls := labelsForGameServer(gs.Name, *gs.Spec.Image)
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      gs.Name,
			Namespace: gs.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &[]int32{1}[0],
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					SecurityContext: &corev1.PodSecurityContext{
						RunAsNonRoot: &[]bool{true}[0],
					},
					Containers: []corev1.Container{
						{
							Name:            "gameserver",
							Image:           *gs.Spec.Image,
							ImagePullPolicy: corev1.PullAlways,
							Ports:           gs.Spec.Ports,
						},
					},
				},
			},
		},
	}

	if err := ctrl.SetControllerReference(gs, dep, r.Scheme); err != nil {
		return nil, err
	}

	return dep, nil
}

func labelsForGameServer(name string, image string) map[string]string {
	var imageTag string
	imageTag = strings.Split(image, ":")[1]

	return map[string]string{
		"app.kubernetes.io/name":                 "gameserver",
		"app.kubernetes.io/instance":             name,
		"app.kubernetes.io/managed-by":           "gameserver-controller",
		"app.kubernetes.io/created-by":           "controller-manager",
		"app.kubernetes.io/part-of":              "gameserver-operator",
		"gameserver.debeijer.io/gameserver-type": imageTag,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *GameServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gameserverv1alpha1.GameServer{}).
		Complete(r)
}
