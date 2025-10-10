package specs

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
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

	sshSidecarEnabled := gs.Spec.SSHSidecar != nil && gs.Spec.SSHSidecar.Enabled

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

	if gs.Spec.Resources != nil {
		resources := corev1ac.ResourceRequirements()
		if gs.Spec.Resources.Limits != nil {
			resources.WithLimits(gs.Spec.Resources.Limits)
		}
		if gs.Spec.Resources.Requests != nil {
			resources.WithRequests(gs.Spec.Resources.Requests)
		}
		container.WithResources(resources)
	}

	if gs.Spec.Service != nil && len(gs.Spec.Service.Ports) > 0 {
		for _, port := range gs.Spec.Service.Ports {
			containerPort := port.Port
			if port.TargetPort.IntVal != 0 {
				containerPort = port.TargetPort.IntVal
			} else if port.TargetPort.StrVal != "" {
				continue
			}

			container.WithPorts(
				corev1ac.ContainerPort().
					WithName(port.Name).
					WithContainerPort(containerPort).
					WithProtocol(port.Protocol),
			)
		}
	}

	if storageEnabled {
		container.WithVolumeMounts(
			corev1ac.VolumeMount().
				WithName("data").
				WithMountPath("/data"),
		)
	}

	// Add shared volume for SSH sidecar
	if sshSidecarEnabled {
		container.WithVolumeMounts(
			corev1ac.VolumeMount().
				WithName("shared").
				WithMountPath("/shared"),
		)
	}

	podSpec := corev1ac.PodSpec().
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
		WithContainers(container)

	// Add SSH sidecar container if enabled
	if sshSidecarEnabled {
		sshImage := "linuxserver/openssh-server:latest"
		if gs.Spec.SSHSidecar.Image != "" {
			sshImage = gs.Spec.SSHSidecar.Image
		}

		sshPort := int32(2222)
		if gs.Spec.SSHSidecar.Port != 0 {
			sshPort = gs.Spec.SSHSidecar.Port
		}

		sshContainer := corev1ac.Container().
			WithName("ssh-sidecar").
			WithImage(sshImage).
			WithImagePullPolicy(v1.PullIfNotPresent).
			WithPorts(
				corev1ac.ContainerPort().
					WithName("ssh").
					WithContainerPort(sshPort).
					WithProtocol(v1.ProtocolTCP),
			).
			WithVolumeMounts(
				corev1ac.VolumeMount().
					WithName("shared").
					WithMountPath("/shared"),
			).
			WithEnv(
				corev1ac.EnvVar().
					WithName("PUID").
					WithValue("1000"),
				corev1ac.EnvVar().
					WithName("PGID").
					WithValue("1000"),
				corev1ac.EnvVar().
					WithName("TZ").
					WithValue("Etc/UTC"),
				corev1ac.EnvVar().
					WithName("USER_NAME").
					WithValue("gameserver"),
			).
			WithSecurityContext(corev1ac.SecurityContext().
				WithAllowPrivilegeEscalation(false).
				WithCapabilities(corev1ac.Capabilities().
					WithDrop("ALL"),
				),
			)

		if len(gs.Spec.SSHSidecar.PublicKeys) > 0 {
			publicKeysStr := ""
			for _, key := range gs.Spec.SSHSidecar.PublicKeys {
				publicKeysStr += key + "\n"
			}
			sshContainer.WithEnv(
				corev1ac.EnvVar().
					WithName("PUBLIC_KEY").
					WithValue(publicKeysStr),
			)
		} else {
			sshContainer.WithEnv(
				corev1ac.EnvVar().
					WithName("PASSWORD_ACCESS").
					WithValue("true"),
				corev1ac.EnvVar().
					WithName("USER_PASSWORD").
					WithValue("changeme"),
			)
		}

		podSpec.WithContainers(sshContainer)
		podSpec.WithVolumes(
			corev1ac.Volume().
				WithName("shared").
				WithEmptyDir(corev1ac.EmptyDirVolumeSource()),
		)
	}

	// Force replica to be 1 or 0
	replicaCount := int32(1)
	if gs.Spec.Replicas == 0 {
		replicaCount = 0
	}

	stsSpec := appsv1ac.StatefulSetSpec().
		WithReplicas(replicaCount).
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
				WithSpec(podSpec),
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

func BuildGameServerService(gs *gamesv1alpha1.GameServer) *corev1ac.ServiceApplyConfiguration {
	if gs.Spec.Service == nil {
		return nil
	}

	servicePorts := make([]*corev1ac.ServicePortApplyConfiguration, 0, len(gs.Spec.Service.Ports))
	for _, port := range gs.Spec.Service.Ports {
		targetPort := port.TargetPort
		if targetPort.IntVal == 0 && targetPort.StrVal == "" {
			targetPort = intstr.FromInt32(port.Port)
		}

		servicePort := corev1ac.ServicePort().
			WithName(port.Name).
			WithProtocol(port.Protocol).
			WithPort(port.Port).
			WithTargetPort(targetPort)

		if port.NodePort != 0 {
			servicePort.WithNodePort(port.NodePort)
		}

		servicePorts = append(servicePorts, servicePort)
	}

	svc := corev1ac.Service(gs.Name, gs.Namespace).
		WithLabels(map[string]string{
			"app.kubernetes.io/name":       utils.GameServerOperatorName,
			"app.kubernetes.io/instance":   gs.Name,
			"app.kubernetes.io/managed-by": utils.GameServerControllerName,
		}).
		WithAnnotations(gs.Spec.Service.Annotations).
		WithSpec(corev1ac.ServiceSpec().
			WithType(gs.Spec.Service.Type).
			WithSelector(map[string]string{
				"app.kubernetes.io/name":       utils.GameServerOperatorName,
				"app.kubernetes.io/instance":   gs.Name,
				"app.kubernetes.io/managed-by": utils.GameServerControllerName,
			}).
			WithPorts(servicePorts...),
		)

	return svc
}
