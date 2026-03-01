package specs_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"

	gamesv1alpha1 "github.com/idebeijer/gameserver-operator/api/v1alpha1"
	"github.com/idebeijer/gameserver-operator/pkg/specs"
	"github.com/idebeijer/gameserver-operator/pkg/utils"
)

var _ = Describe("LinuxGSM spec builders", func() {
	Describe("BuildLinuxGSMGameServerStatefulSet", func() {
		It("builds a stateful set with persistent storage and secure defaults", func() {
			gs := newGameServer()

			statefulSet := specs.BuildLinuxGSMGameServerStatefulSet(gs)
			Expect(statefulSet).NotTo(BeNil())
			Expect(statefulSet.Spec).NotTo(BeNil())

			Expect(statefulSet.Spec.Replicas).NotTo(BeNil())
			Expect(*statefulSet.Spec.Replicas).To(Equal(int32(1)))

			expectedLabels := expectedLabels(gs)
			Expect(statefulSet.Spec.Selector).NotTo(BeNil())
			Expect(statefulSet.Spec.Selector.MatchLabels).To(Equal(expectedLabels))

			Expect(statefulSet.Spec.Template).NotTo(BeNil())
			Expect(statefulSet.Spec.Template.Labels).To(Equal(expectedLabels))
			Expect(statefulSet.Spec.Template.Spec).NotTo(BeNil())

			podSpec := statefulSet.Spec.Template.Spec
			Expect(podSpec.SecurityContext).NotTo(BeNil())
			Expect(podSpec.Containers).To(HaveLen(1))
			Expect(podSpec.Containers[0].SecurityContext).NotTo(BeNil())
			Expect(podSpec.SecurityContext.RunAsNonRoot).To(HaveValue(BeTrue()))
			Expect(podSpec.SecurityContext.RunAsUser).To(HaveValue(Equal(int64(1000))))
			Expect(podSpec.SecurityContext.RunAsGroup).To(HaveValue(Equal(int64(1000))))
			Expect(podSpec.SecurityContext.FSGroup).To(HaveValue(Equal(int64(1000))))
			Expect(podSpec.AutomountServiceAccountToken).To(HaveValue(BeFalse()))

			container := podSpec.Containers[0]
			Expect(container.Name).NotTo(BeNil())
			Expect(*container.Name).To(Equal("gameserver"))
			Expect(container.Image).NotTo(BeNil())
			Expect(*container.Image).To(Equal("gameservermanagers/gameserver:valheim"))
			Expect(container.Command).To(Equal([]string{"/app/entrypoint-user.sh"}))

			Expect(container.Env).To(HaveLen(1))
			env := container.Env[0]
			Expect(env.Name).NotTo(BeNil())
			Expect(*env.Name).To(Equal("UPDATE_CHECK"))
			Expect(env.Value).NotTo(BeNil())
			Expect(*env.Value).To(Equal("0"))

			Expect(container.VolumeMounts).To(HaveLen(1))
			volumeMount := container.VolumeMounts[0]
			Expect(volumeMount.Name).NotTo(BeNil())
			Expect(*volumeMount.Name).To(Equal("data"))
			Expect(volumeMount.MountPath).NotTo(BeNil())
			Expect(*volumeMount.MountPath).To(Equal("/data"))

			Expect(statefulSet.Spec.VolumeClaimTemplates).To(HaveLen(1))
			pvc := statefulSet.Spec.VolumeClaimTemplates[0]
			Expect(pvc.Name).NotTo(BeNil())
			Expect(*pvc.Name).To(Equal("data"))
			Expect(pvc.Spec).NotTo(BeNil())
			Expect(pvc.Spec.AccessModes).To(Equal([]corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}))
			Expect(pvc.Spec.Resources).NotTo(BeNil())
			Expect(pvc.Spec.Resources.Requests).NotTo(BeNil())

			requests := *pvc.Spec.Resources.Requests
			Expect(requests).To(HaveKey(corev1.ResourceStorage))
			quantity := requests[corev1.ResourceStorage]
			Expect((&quantity).Cmp(resource.MustParse("10Gi"))).To(Equal(0))
		})

		It("omits persistent storage when storage is disabled", func() {
			gs := newGameServer(func(gs *gamesv1alpha1.GameServer) {
				gs.Spec.Storage = &gamesv1alpha1.StorageSpec{
					Enabled: new(false),
				}
			})

			statefulSet := specs.BuildLinuxGSMGameServerStatefulSet(gs)
			Expect(statefulSet).NotTo(BeNil())
			Expect(statefulSet.Spec).NotTo(BeNil())
			Expect(statefulSet.Spec.Template).NotTo(BeNil())
			Expect(statefulSet.Spec.Template.Spec).NotTo(BeNil())

			podSpec := statefulSet.Spec.Template.Spec
			Expect(podSpec.Containers).To(HaveLen(1))
			Expect(podSpec.Containers[0].VolumeMounts).To(BeEmpty())
			Expect(statefulSet.Spec.VolumeClaimTemplates).To(BeNil())
		})

		It("adds container ports for numeric service target ports", func() {
			gs := newGameServer(func(gs *gamesv1alpha1.GameServer) {
				gs.Spec.Service = &gamesv1alpha1.ServiceSpec{
					Ports: []gamesv1alpha1.ServicePort{
						{
							Name:       "primary",
							Port:       27015,
							TargetPort: intstr.FromInt32(28015),
							Protocol:   corev1.ProtocolUDP,
						},
						{
							Name:       "string-port",
							Port:       8080,
							TargetPort: intstr.FromString("metrics"),
							Protocol:   corev1.ProtocolTCP,
						},
					},
				}
			})

			statefulSet := specs.BuildLinuxGSMGameServerStatefulSet(gs)
			Expect(statefulSet.Spec.Template).NotTo(BeNil())
			Expect(statefulSet.Spec.Template.Spec).NotTo(BeNil())

			podSpec := statefulSet.Spec.Template.Spec
			Expect(podSpec.Containers).To(HaveLen(1))
			container := podSpec.Containers[0]

			Expect(container.Ports).To(HaveLen(1))
			port := container.Ports[0]
			Expect(port.Name).NotTo(BeNil())
			Expect(*port.Name).To(Equal("primary"))
			Expect(port.ContainerPort).NotTo(BeNil())
			Expect(*port.ContainerPort).To(Equal(int32(28015)))
			Expect(port.Protocol).NotTo(BeNil())
			Expect(*port.Protocol).To(Equal(corev1.ProtocolUDP))
		})
	})

	Describe("BuildGameServerService", func() {
		It("returns nil when service spec is not provided", func() {
			gs := newGameServer()
			Expect(specs.BuildGameServerService(gs)).To(BeNil())
		})

		It("builds a service with labels, annotations, and default target ports", func() {
			gs := newGameServer(func(gs *gamesv1alpha1.GameServer) {
				gs.Spec.Service = &gamesv1alpha1.ServiceSpec{
					Type: corev1.ServiceTypeNodePort,
					Annotations: map[string]string{
						"service.beta.kubernetes.io/aws-load-balancer-type": "nlb",
					},
					Ports: []gamesv1alpha1.ServicePort{
						{
							Name:     "primary",
							Protocol: corev1.ProtocolTCP,
							Port:     25565,
							NodePort: 30010,
						},
						{
							Name:       "metrics",
							Protocol:   corev1.ProtocolUDP,
							Port:       8125,
							TargetPort: intstr.FromString("udp-metrics"),
						},
					},
				}
			})

			service := specs.BuildGameServerService(gs)
			Expect(service).NotTo(BeNil())
			Expect(service.Labels).To(Equal(expectedLabels(gs)))
			Expect(service.Annotations).To(HaveKeyWithValue("service.beta.kubernetes.io/aws-load-balancer-type", "nlb"))

			Expect(service.Spec).NotTo(BeNil())
			Expect(service.Spec.Type).NotTo(BeNil())
			Expect(*service.Spec.Type).To(Equal(corev1.ServiceTypeNodePort))
			Expect(service.Spec.Selector).To(Equal(expectedLabels(gs)))
			Expect(service.Spec.Ports).To(HaveLen(2))

			primary := service.Spec.Ports[0]
			Expect(primary.Name).NotTo(BeNil())
			Expect(*primary.Name).To(Equal("primary"))
			Expect(primary.Port).NotTo(BeNil())
			Expect(*primary.Port).To(Equal(int32(25565)))
			Expect(primary.TargetPort).NotTo(BeNil())
			Expect(primary.TargetPort.IntVal).To(Equal(int32(25565)))
			Expect(primary.NodePort).NotTo(BeNil())
			Expect(*primary.NodePort).To(Equal(int32(30010)))

			metrics := service.Spec.Ports[1]
			Expect(metrics.Name).NotTo(BeNil())
			Expect(*metrics.Name).To(Equal("metrics"))
			Expect(metrics.TargetPort).NotTo(BeNil())
			Expect(metrics.TargetPort.StrVal).To(Equal("udp-metrics"))
		})

		Describe("Traffic Policy Configuration", func() {
			It("omits traffic policies when not set in GameServer spec", func() {
				gs := newGameServer(func(gs *gamesv1alpha1.GameServer) {
					gs.Spec.Service = &gamesv1alpha1.ServiceSpec{
						Type: corev1.ServiceTypeLoadBalancer,
						Ports: []gamesv1alpha1.ServicePort{
							{
								Name:     "game",
								Protocol: corev1.ProtocolTCP,
								Port:     25565,
							},
						},
						// ExternalTrafficPolicy and InternalTrafficPolicy not set (nil)
					}
				})

				service := specs.BuildGameServerService(gs)
				Expect(service).NotTo(BeNil())
				Expect(service.Spec.ExternalTrafficPolicy).To(BeNil())
				Expect(service.Spec.InternalTrafficPolicy).To(BeNil())
			})

			It("sets both traffic policies when specified", func() {
				gs := newGameServer(func(gs *gamesv1alpha1.GameServer) {
					gs.Spec.Service = &gamesv1alpha1.ServiceSpec{
						Type:                  corev1.ServiceTypeLoadBalancer,
						ExternalTrafficPolicy: ptr.To(corev1.ServiceExternalTrafficPolicyLocal),
						InternalTrafficPolicy: ptr.To(corev1.ServiceInternalTrafficPolicyCluster),
						Ports: []gamesv1alpha1.ServicePort{
							{
								Name:     "game",
								Protocol: corev1.ProtocolTCP,
								Port:     25565,
							},
						},
					}
				})

				service := specs.BuildGameServerService(gs)
				Expect(service).NotTo(BeNil())
				Expect(service.Spec.ExternalTrafficPolicy).NotTo(BeNil())
				Expect(*service.Spec.ExternalTrafficPolicy).To(Equal(corev1.ServiceExternalTrafficPolicyLocal))
				Expect(service.Spec.InternalTrafficPolicy).NotTo(BeNil())
				Expect(*service.Spec.InternalTrafficPolicy).To(Equal(corev1.ServiceInternalTrafficPolicyCluster))
			})
		})
	})
})

func newGameServer(overrides ...func(*gamesv1alpha1.GameServer)) *gamesv1alpha1.GameServer {
	gs := &gamesv1alpha1.GameServer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example",
			Namespace: "default",
		},
		Spec: gamesv1alpha1.GameServerSpec{
			GameName: "valheim",
			Replicas: 1,
		},
	}

	for _, override := range overrides {
		override(gs)
	}

	return gs
}

func expectedLabels(gs *gamesv1alpha1.GameServer) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":       utils.GameServerOperatorName,
		"app.kubernetes.io/instance":   gs.Name,
		"app.kubernetes.io/managed-by": utils.GameServerControllerName,
	}
}
