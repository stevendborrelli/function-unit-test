// Package v1beta1 contains the input type for this Function
// +kubebuilder:object:generate=true
// +groupName=unittest.fn.crossplane.io
// +versionName=v1beta1
package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Test struct {
	// Description is a description of the test
	Description string `json:"description,omitempty"`
	// Test is a CEL expression to evaluate. If true the test passes,
	// if false the test fails
	Assert string `json:"assert"`
}

// Input can be used to provide input to this Function.
// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +kubebuilder:resource:categories=crossplane
type Input struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// ErrorOnFailedTest whether we return an error if any test fails
	// default is false
	ErrorOnFailedTest bool `json:"errorOnFailedTest"`

	// Example is an example field. Replace it with whatever input you need. :)
	TestCases []Test `json:"testCases"`
}
