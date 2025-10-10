package specs

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	appsv1ac "k8s.io/client-go/applyconfigurations/apps/v1"
	corev1ac "k8s.io/client-go/applyconfigurations/core/v1"
	metav1ac "k8s.io/client-go/applyconfigurations/meta/v1"

	gamesv1alpha1 "github.com/idebeijer/gameserver-operator/api/v1alpha1"
	"github.com/idebeijer/gameserver-operator/pkg/utils"
)

func BuildLinuxGSMGameServerStatefulSet(gs *gamesv1alpha1.GameServer) *appsv1ac.StatefulSetApplyConfiguration {
	image := fmt.Sprintf("gameservermanagers/gameserver:%s", gs.Spec.GameName)

	storageEnabled := true
	if gs.Spec.Storage != nil && gs.Spec.Storage.Enabled != nil {
		storageEnabled = *gs.Spec.Storage.Enabled
	}

	container := corev1ac.Container().
		WithName("gameserver").
		WithImage(image).
		WithImagePullPolicy(v1.PullIfNotPresent).
		WithSecurityContext(corev1ac.SecurityContext().
			WithAllowPrivilegeEscalation(false).
			WithCapabilities(corev1ac.Capabilities().
				WithDrop("ALL"),
			),
		).
		WithCommand("/app/entrypoint-user.sh").
		WithEnv(
			corev1ac.EnvVar().
				WithName("UPDATE_CHECK").
				WithValue("0"),
		)

	if storageEnabled {
		container.WithVolumeMounts(
			corev1ac.VolumeMount().
				WithName("data").
				WithMountPath("/data"),
		)
	}

	stsSpec := appsv1ac.StatefulSetSpec().
		WithReplicas(1).
		WithSelector(metav1ac.LabelSelector().
			WithMatchLabels(map[string]string{
				"app.kubernetes.io/name":       utils.GameServerOperatorName,
				"app.kubernetes.io/instance":   gs.Name,
				"app.kubernetes.io/managed-by": utils.GameServerControllerName,
			}),
		).
		WithTemplate(
			corev1ac.PodTemplateSpec().
				WithLabels(map[string]string{
					"app.kubernetes.io/name":       utils.GameServerOperatorName,
					"app.kubernetes.io/instance":   gs.Name,
					"app.kubernetes.io/managed-by": utils.GameServerControllerName,
				}).
				WithSpec(corev1ac.PodSpec().
					WithSecurityContext(corev1ac.PodSecurityContext().
						WithRunAsNonRoot(true).
						WithRunAsUser(1000).
						WithRunAsGroup(1000).
						WithFSGroup(1000).
						WithFSGroupChangePolicy(v1.FSGroupChangeOnRootMismatch).
						WithSeccompProfile(corev1ac.SeccompProfile().
							WithType(v1.SeccompProfileTypeRuntimeDefault),
						),
					).
					WithContainers(container),
				),
		)

	if storageEnabled {
		storageSize := "10Gi"
		if gs.Spec.Storage != nil && gs.Spec.Storage.Size != "" {
			storageSize = gs.Spec.Storage.Size
		}

		pvcSpec := corev1ac.PersistentVolumeClaimSpec().
			WithAccessModes(v1.ReadWriteOnce).
			WithResources(corev1ac.VolumeResourceRequirements().
				WithRequests(v1.ResourceList{
					v1.ResourceStorage: resource.MustParse(storageSize),
				}),
			)

		if gs.Spec.Storage != nil && gs.Spec.Storage.StorageClassName != nil {
			pvcSpec.WithStorageClassName(*gs.Spec.Storage.StorageClassName)
		}

		stsSpec.WithVolumeClaimTemplates(
			corev1ac.PersistentVolumeClaim("data", gs.Namespace).
				WithSpec(pvcSpec),
		)
	}

	sts := appsv1ac.StatefulSet(gs.Name, gs.Namespace).
		WithLabels(map[string]string{}).
		WithAnnotations(map[string]string{}).
		WithSpec(stsSpec)

	return sts
}
