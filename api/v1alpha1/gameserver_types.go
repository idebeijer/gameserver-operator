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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GameServerSpec defines the desired state of GameServer
// leave out for now //+kubebuilder:validation:XValidation:rule="size(self.gameConfigs) <= 1",message="Cannot specify more than one game-specific configuration block."
type GameServerSpec struct {
	// GameName is the name of the game server.
	// For LinuxGSM, this should match the shortname of a supported game server.
	// Examples include 'rust' for Rust and 'mc' for Minecraft.
	//
	// For a list of supported games, see:
	// https://github.com/GameServerManagers/LinuxGSM/blob/master/lgsm/data/serverlist.csv
	// +kubebuilder:validation:Required
	GameName string `json:"gameName,omitempty"`

	// Manager specifies the installation and management tool for the game server.
	// 'LinuxGSM' is the default and currently the only supported option.
	// +kubebuilder:validation:Enum=LinuxGSM
	// +kubebuilder:default=LinuxGSM
	// +optional
	Manager string `json:"manager,omitempty"`

	// GameVersion is the version of the game server.
	// +optional
	GameVersion string `json:"gameVersion,omitempty"`

	// GameConfigs holds game-specific configuration options.
	// +optional
	GameConfigs *GameConfigs `json:"gameConfigs,omitempty"`

	// Replicas is the number of game server instances to run.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=1
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// Storage defines the storage configuration for the game server.
	// If not specified, a default storage size of 10Gi will be used.
	// +optional
	Storage *StorageSpec `json:"storage,omitempty"`
}

type GameConfigs struct {
	// Minecraft holds configuration specific to Minecraft game servers.
	// +optional
	Minecraft *MinecraftConfig `json:"minecraft,omitempty"`
}

type MinecraftConfig struct {
	// Version specifies the Minecraft server version.
	// +optional
	Version string `json:"version,omitempty"`

	// Mods is a list of mods to be installed on the Minecraft server.
	// +optional
	Mods []string `json:"mods,omitempty"`
}

// StorageSpec defines the storage configuration for the game server.
type StorageSpec struct {
	// Enabled indicates whether persistent storage is enabled for the game server.
	// If not specified, storage is enabled by default.
	// +kubebuilder:default=true
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// Size is the size of the persistent volume claim for the game server data.
	// +kubebuilder:validation:Pattern=`^\d+Gi$`
	// +kubebuilder:default="10Gi"
	// +optional
	Size string `json:"size,omitempty"`

	// StorageClassName is the name of the StorageClass to use for the persistent volume claim.
	// If not specified, the default StorageClass for the cluster will be used.
	// +optional
	StorageClassName *string `json:"storageClassName,omitempty"`
}

// GameServerStatus defines the observed state of GameServer.
type GameServerStatus struct {
	// For Kubernetes API conventions, see:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties

	// conditions represent the current state of the GameServer resource.
	// Each condition has a unique type and reflects the status of a specific aspect of the resource.
	//
	// Standard condition types include:
	// - "Available": the resource is fully functional
	// - "Progressing": the resource is being created or updated
	// - "Degraded": the resource failed to reach or maintain its desired state
	//
	// The status of each condition is one of True, False, or Unknown.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// GameServer is the Schema for the gameservers API
type GameServer struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of GameServer
	// +required
	Spec GameServerSpec `json:"spec"`

	// status defines the observed state of GameServer
	// +optional
	Status GameServerStatus `json:"status,omitempty,omitzero"`
}

// +kubebuilder:object:root=true

// GameServerList contains a list of GameServer
type GameServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GameServer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GameServer{}, &GameServerList{})
}
