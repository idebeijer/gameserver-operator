package specs

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	appsv1ac "k8s.io/client-go/applyconfigurations/apps/v1"
	corev1ac "k8s.io/client-go/applyconfigurations/core/v1"
	metav1ac "k8s.io/client-go/applyconfigurations/meta/v1"

	gamesv1alpha1 "github.com/idebeijer/gameserver-operator/api/v1alpha1"
	"github.com/idebeijer/gameserver-operator/pkg/utils"
)

func BuildLinuxGSMGameServerStatefulSet(gs *gamesv1alpha1.GameServer) *appsv1ac.StatefulSetApplyConfiguration {
	image := fmt.Sprintf("gameservermanagers/gameserver:%s", gs.Spec.GameName)
	sts := appsv1ac.StatefulSet(gs.Name, gs.Namespace).
		WithLabels(map[string]string{}).
		WithAnnotations(map[string]string{}).
		WithSpec(appsv1ac.StatefulSetSpec().
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
						WithContainers(
							corev1ac.Container().
								WithName("gameserver").
								WithImage(image).
								WithImagePullPolicy(v1.PullIfNotPresent).
								WithSecurityContext(corev1ac.SecurityContext().
									WithAllowPrivilegeEscalation(false).
									WithCapabilities(corev1ac.Capabilities().
										WithDrop("ALL"),
									),
								).
								// Skip entrypoint.sh and use the user entrypoint to avoid running as root.
								// That means cron is not started, so update checks are disabled.
								// TODO: Add a sidecar container that runs cron and does update checks.
								WithCommand("/app/entrypoint-user.sh").
								WithEnv(
									corev1ac.EnvVar().
										WithName("UPDATE_CHECK").
										WithValue("0"),
								),
						),
					),
			),
		)

	return sts
}
