package controller

import (
	"context"

	metav1ac "k8s.io/client-go/applyconfigurations/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	gamesv1alpha1 "github.com/idebeijer/gameserver-operator/api/v1alpha1"
	"github.com/idebeijer/gameserver-operator/pkg/specs"
)

func (r *GameServerReconciler) reconcileGameServer(ctx context.Context, gs *gamesv1alpha1.GameServer) error {
	if err := r.reconcileGameServerStatefulSet(ctx, gs); err != nil {
		return err
	}

	if err := r.reconcileGameServerService(ctx, gs); err != nil {
		return err
	}

	return nil
}

func (r *GameServerReconciler) reconcileGameServerStatefulSet(ctx context.Context, gs *gamesv1alpha1.GameServer) error {
	switch gs.Spec.Manager {
	case "LinuxGSM":
		return r.reconcileLinuxGSMGameServerStatefulSet(ctx, gs)
	default:
		return nil
	}
}

func (r *GameServerReconciler) reconcileLinuxGSMGameServerStatefulSet(ctx context.Context, gs *gamesv1alpha1.GameServer) error {
	stsApply := specs.BuildLinuxGSMGameServerStatefulSet(gs)
	stsApply.WithOwnerReferences(metav1ac.OwnerReference().
		WithAPIVersion(gs.APIVersion).
		WithKind(gs.Kind).
		WithName(gs.Name).
		WithUID(gs.UID).
		WithController(true).
		WithBlockOwnerDeletion(true),
	)

	if err := r.Apply(ctx, stsApply,
		client.FieldOwner(fieldManagerGameServer),
		client.ForceOwnership,
	); err != nil {
		return err
	}

	return nil
}

func (r *GameServerReconciler) reconcileGameServerService(ctx context.Context, gs *gamesv1alpha1.GameServer) error {
	if gs.Spec.Service == nil {
		return nil
	}

	svcApply := specs.BuildGameServerService(gs)
	svcApply.WithOwnerReferences(metav1ac.OwnerReference().
		WithAPIVersion(gs.APIVersion).
		WithKind(gs.Kind).
		WithName(gs.Name).
		WithUID(gs.UID).
		WithController(true).
		WithBlockOwnerDeletion(true),
	)

	if err := r.Apply(ctx, svcApply,
		client.FieldOwner(fieldManagerGameServer),
		client.ForceOwnership,
	); err != nil {
		return err
	}

	return nil
}
