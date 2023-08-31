/*
Copyright 2022 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// CanaryParameters are the configurable fields of a Canary.
//type CanaryParameters struct {
//	//ConfigurableField string `json:"configurableField"`
//	Name   string `json:"name"`
//	Weight string `json:"weight"`
//	Destination string `json:"destination"`
//}

type Dst struct {
    Service string `json:"service"`
    Version string `json:"version"`
}

type  Abc struct {
  Weight int `json:"weight"`
  Destination Dst `json:"destination"`
}
  //Destination struct {
  //  Service string `json:"service"`
  //  Version string `json:"version"`
  //} `json:"destination"`

type CanaryParameters struct {
	Sources []string `json:"sources"`
        Split [] Abc `json:"split"`

}
/*
type CanaryParameters struct {
	Sources []string `json:"sources"`
	Conf    struct {
		Split []struct {
			Weight      int `json:"weight"`
			Destination struct {
				Service string `json:"service"`
				Version string `json:"version"`
			} `json:"destination"`
		} `json:"split"`
	} `json:"conf"`
}
*/
// CanaryObservation are the observable fields of a Canary.
type CanaryObservation struct {
	ObservableField string `json:"observableField,omitempty"`
}

// A CanarySpec defines the desired state of a Canary.
type CanarySpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       CanaryParameters `json:"forProvider"`
}

// A CanaryStatus represents the observed state of a Canary.
type CanaryStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          CanaryObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Canary is an example API type.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,istio}
type Canary struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec   CanarySpec   `json:"spec"`
	Status CanaryStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CanaryList contains a list of Canary
type CanaryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Canary `json:"items"`
}

// Canary type metadata.
var (
	CanaryKind             = reflect.TypeOf(Canary{}).Name()
	CanaryGroupKind        = schema.GroupKind{Group: Group, Kind: CanaryKind}.String()
	CanaryKindAPIVersion   = CanaryKind + "." + SchemeGroupVersion.String()
	CanaryGroupVersionKind = SchemeGroupVersion.WithKind(CanaryKind)
)

func init() {
	SchemeBuilder.Register(&Canary{}, &CanaryList{})
}
