/*
The MIT License (MIT)

Copyright © 2025 Igor de Beijer

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package controller

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	gamesv1alpha1 "github.com/idebeijer/gameserver-operator/api/v1alpha1"
)

const (
	finalizerName          = "gameserver.games.idebeijer.github.io/finalizer"
	fieldManagerGameServer = "gameserver-controller"
)

// GameServerReconciler reconciles a GameServer object
type GameServerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=games.idebeijer.github.io,resources=gameservers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=games.idebeijer.github.io,resources=gameservers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=games.idebeijer.github.io,resources=gameservers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *GameServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = logf.FromContext(ctx)

	gs, err := r.getGameServer(ctx, req)
	if err != nil {
		return ctrl.Result{}, err
	}
	if gs == nil {
		// GameServer resource not found. Ignoring since object must be deleted.
		return ctrl.Result{}, nil
	}

	if err := r.reconcileGameServer(ctx, gs); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GameServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gamesv1alpha1.GameServer{}).
		Named("gameserver").
		Owns(&corev1.Service{}).
		Owns(&appsv1.StatefulSet{}).
		Complete(r)
}

func (r *GameServerReconciler) getGameServer(ctx context.Context, req ctrl.Request) (*gamesv1alpha1.GameServer, error) {
	gs := &gamesv1alpha1.GameServer{}
	if err := r.Get(ctx, req.NamespacedName, gs); err != nil {
		if apierrors.IsNotFound(err) {
			// Object not found, return. Created objects are automatically garbage collected.
			return nil, nil
		}
		// Error reading the object, requeue the request.
		return nil, fmt.Errorf("failed to get GameServer: %w", err)
	}

	// Ensure TypeMeta is set for owner references
	// The API server doesn't always populate these fields
	// TODO: check if this is actually needed or an issue? (lacking support in k8s or fake client?)
	if gs.APIVersion == "" {
		gs.APIVersion = gamesv1alpha1.GroupVersion.String()
	}
	if gs.Kind == "" {
		gs.Kind = "GameServer"
	}

	return gs, nil
}
