package controller

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1ac "k8s.io/client-go/applyconfigurations/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	gamesv1alpha1 "github.com/idebeijer/gameserver-operator/api/v1alpha1"
	"github.com/idebeijer/gameserver-operator/pkg/specs"
)

func (r *GameServerReconciler) reconcileGameServer(ctx context.Context, gs *gamesv1alpha1.GameServer) error {
	if err := r.reconcileEditorSecret(ctx, gs); err != nil {
		return err
	}

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

// reconcileEditorSecret ensures an auto-generated password Secret exists when the editor
// sidecar is enabled with password auth and no external secret is referenced.
// The Secret is only created, never updated, so the password survives reconcile loops.
func (r *GameServerReconciler) reconcileEditorSecret(ctx context.Context, gs *gamesv1alpha1.GameServer) error {
	editor := gs.Spec.Editor
	if editor == nil || !editor.Enabled {
		return nil
	}
	if editor.Auth != nil && ((editor.Auth.Enabled != nil && !*editor.Auth.Enabled) || editor.Auth.PasswordSecretRef != nil) {
		return nil
	}

	secretName := specs.EditorPasswordSecretName(gs)
	existing := &corev1.Secret{}
	err := r.Get(ctx, client.ObjectKey{Namespace: gs.Namespace, Name: secretName}, existing)
	if err == nil {
		return nil
	}
	if !apierrors.IsNotFound(err) {
		return fmt.Errorf("failed to get editor Secret: %w", err)
	}

	password, err := generatePassword(24)
	if err != nil {
		return fmt.Errorf("failed to generate editor password: %w", err)
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: gs.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         gs.APIVersion,
					Kind:               gs.Kind,
					Name:               gs.Name,
					UID:                gs.UID,
					Controller:         &[]bool{true}[0],
					BlockOwnerDeletion: &[]bool{true}[0],
				},
			},
		},
		StringData: map[string]string{
			"password": password,
		},
	}

	if err := r.Create(ctx, secret); err != nil && !apierrors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create editor Secret: %w", err)
	}

	return nil
}

const passwordChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generatePassword(length int) (string, error) {
	b := make([]byte, length)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(passwordChars))))
		if err != nil {
			return "", err
		}
		b[i] = passwordChars[n.Int64()]
	}
	return string(b), nil
}
