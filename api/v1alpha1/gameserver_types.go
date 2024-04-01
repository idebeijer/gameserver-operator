package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type LinuxGSMConfigSource struct {
	// Config is the LinuxGSM config
	Config *string `json:"config,omitempty"`

	// ExistingConfigMap is a reference to a ConfigMap containing the LinuxGSM config
	// If this is set, Config is ignored
	ExistingConfigMap *v1.ConfigMapKeySelector `json:"existingConfigMap,omitempty"`
}

type GameConfigSource struct {
	//// ConfigFileName will be the name of the file to write the config to
	//ConfigFileName *string `json:"configFileName,omitempty"`

	// Config is the game config
	Config *string `json:"config,omitempty"`

	// ExistingConfigMap is a reference to a ConfigMap containing the game config
	// If this is set, Config is ignored
	ExistingConfigMap *v1.ConfigMapKeySelector `json:"existingConfigMap,omitempty"`

	// MountPath is the path to mount the config file
	MountPath *string `json:"mountPath,omitempty"`
}

type GameServerDataPVC struct {
	// Enabled is a flag to enable or disable the PVC
	// +kubebuilder:default=true
	Enabled *bool `json:"enabled,omitempty"`

	// Name is the name of the PVC
	Name *string `json:"name,omitempty"`

	// Resources represents the minimum resources the volume should have.
	// More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#resources
	Resources v1.ResourceRequirements `json:"resources,omitempty"`

	// StorageClassName is the name of the StorageClass required by the claim.
	// More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#class-1
	StorageClassName *string `json:"storageClassName,omitempty"`

	// Selector is the label selector for the PVC
	Selector *metav1.LabelSelector `json:"selector,omitempty"`
}

// GameServerSpec defines the desired state of GameServer
type GameServerSpec struct {

	// Image is the container image to run
	Image *string `json:"image"`

	// LinuxGSMConfig is the LinuxGSM configuration
	LinuxGSMConfig *LinuxGSMConfigSource `json:"linuxGSMConfig,omitempty"`

	// GameConfigs is a list of game configs
	GameConfigs []GameConfigSource `json:"gameConfigs,omitempty"`

	// Ports is a list of ports to expose
	Ports []v1.ContainerPort `json:"ports,omitempty"`

	// GameServerDataPVC is the spec for the persistent volume claim
	GameServerDataPVC *GameServerDataPVC `json:"gameServerDataPVC,omitempty"`

	// UseHostNetwork is a flag to use the host network of the node (recommended for game servers)
	// +kubebuilder:default=true
	UseHostNetwork *bool `json:"useHostNetwork,omitempty"`

	// AdditionalVolumes is a list of additional volumes to mount
	AdditionalVolumes []v1.Volume `json:"additionalVolumes,omitempty"`

	// AdditionalVolumeMounts is a list of additional volume mounts
	AdditionalVolumeMounts []v1.VolumeMount `json:"additionalVolumeMounts,omitempty"`
}

// GameServerStatus defines the observed state of GameServer
type GameServerStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// GameServer is the Schema for the gameservers API
type GameServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GameServerSpec   `json:"spec,omitempty"`
	Status GameServerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// GameServerList contains a list of GameServer
type GameServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GameServer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GameServer{}, &GameServerList{})
}
