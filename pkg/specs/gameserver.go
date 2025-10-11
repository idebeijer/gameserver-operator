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
	storageEnabled := linuxGSMStorageEnabled(gs)
	container := buildLinuxGSMContainer(gs, storageEnabled)
	podSpec := buildLinuxGSMPodSpec(container)
	stsSpec := buildLinuxGSMStatefulSetSpec(gs, podSpec, storageEnabled)

	sts := appsv1ac.StatefulSet(gs.Name, gs.Namespace).
		WithLabels(map[string]string{}).
		WithAnnotations(map[string]string{}).
		WithSpec(stsSpec)

	return sts
}

func linuxGSMStorageEnabled(gs *gamesv1alpha1.GameServer) bool {
	if gs.Spec.Storage != nil && gs.Spec.Storage.Enabled != nil {
		return *gs.Spec.Storage.Enabled
	}
	return true
}

func buildLinuxGSMContainer(gs *gamesv1alpha1.GameServer, storageEnabled bool) *corev1ac.ContainerApplyConfiguration {
	container := corev1ac.Container().
		WithName("gameserver").
		WithImage(fmt.Sprintf("gameservermanagers/gameserver:%s", gs.Spec.GameName)).
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

	if gs.Spec.Service != nil {
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

	return container
}

func buildLinuxGSMPodSpec(container *corev1ac.ContainerApplyConfiguration) *corev1ac.PodSpecApplyConfiguration {
	return corev1ac.PodSpec().
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
}

func buildLinuxGSMStatefulSetSpec(gs *gamesv1alpha1.GameServer, podSpec *corev1ac.PodSpecApplyConfiguration, storageEnabled bool) *appsv1ac.StatefulSetSpecApplyConfiguration {
	replicaCount := int32(1)
	if gs.Spec.Replicas == 0 {
		replicaCount = 0
	}

	stsSpec := appsv1ac.StatefulSetSpec().
		WithReplicas(replicaCount).
		WithSelector(metav1ac.LabelSelector().
			WithMatchLabels(gameServerLabels(gs)),
		).
		WithTemplate(
			corev1ac.PodTemplateSpec().
				WithLabels(gameServerLabels(gs)).
				WithSpec(podSpec),
		)

	if storageEnabled {
		stsSpec.WithVolumeClaimTemplates(buildLinuxGSMVolumeClaimTemplate(gs))
	}

	return stsSpec
}

func buildLinuxGSMVolumeClaimTemplate(gs *gamesv1alpha1.GameServer) *corev1ac.PersistentVolumeClaimApplyConfiguration {
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

	return corev1ac.PersistentVolumeClaim("data", gs.Namespace).
		WithSpec(pvcSpec)
}

func gameServerLabels(gs *gamesv1alpha1.GameServer) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":       utils.GameServerOperatorName,
		"app.kubernetes.io/instance":   gs.Name,
		"app.kubernetes.io/managed-by": utils.GameServerControllerName,
	}
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
		WithLabels(gameServerLabels(gs)).
		WithAnnotations(gs.Spec.Service.Annotations).
		WithSpec(corev1ac.ServiceSpec().
			WithType(gs.Spec.Service.Type).
			WithSelector(gameServerLabels(gs)).
			WithPorts(servicePorts...),
		)

	return svc
}
