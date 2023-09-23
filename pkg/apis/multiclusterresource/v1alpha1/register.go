package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	MultiClusterResourceGroup      = "mulitcluster.practice.com"
	MultiClusterResourceVersion    = "v1alpha1"
	MultiClusterResourceKind       = "MultiClusterResource"
	MultiClusterResourceApiVersion = "mulitcluster.practice.com/v1alpha1"
)

// SchemeGroupVersion is group version used to register these objects
var SchemeGroupVersion = schema.GroupVersion{Group: MultiClusterResourceGroup, Version: MultiClusterResourceVersion}

// Kind takes an unqualified kind and return back a Group qualified GroupKind
func Kind(kind string) schema.GroupKind {
	return SchemeGroupVersion.WithKind(kind).GroupKind()
}

// GetResource takes an unqualified multiclusterresource and returns a Group qualified GroupResource
func GetResource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

var (
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme   = SchemeBuilder.AddToScheme
)

// Adds the list of known types to Scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&MultiClusterResource{},
		&MultiClusterResourceList{},
	)
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
