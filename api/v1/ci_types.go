/*
Copyright 2024.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CISpec defines the desired state of CI
type CISpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of CI. Edit ci_types.go to remove/update
	Repo RepoSpec `json:"repo"`
	// Repo string `json:repo,omitempty`
}

type RepoSpec struct {
	URL            string    `json:"url"`
	AccessToken    string    `json:"accessToken,omitempty"`
	ScrapeInterval int       `json:"scrapeInterval"`
	Jobs           []JobSpec `json:"jobs"`
}

type JobSpec struct {
	Name     string   `json:"name"`
	Stage    string   `json:"stage"`
	Image    string   `json:"image"`
	Commands []string `json:"commands"`
}

type CommandsSpec struct {
	Commands []string `json:"commands"`
}

// CIStatus defines the observed state of CI
type CIStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CI is the Schema for the cis API
type CI struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CISpec   `json:"spec,omitempty"`
	Status CIStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CIList contains a list of CI
type CIList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CI `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CI{}, &CIList{})
}
