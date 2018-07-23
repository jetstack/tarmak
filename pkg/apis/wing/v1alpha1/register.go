// Copyright Jetstack Ltd. See LICENSE for details.
package v1alpha1

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const GroupName = "wing.tarmak.io"

// SchemeGroupVersion is group version used to register these objects
var SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha1"}

var (
	// TODO: move SchemeBuilder with zz_generated.deepcopy.go to k8s.io/api.
	// localSchemeBuilder and AddToScheme will stay in k8s.io/kubernetes.
	SchemeBuilder      runtime.SchemeBuilder
	localSchemeBuilder = &SchemeBuilder
	AddToScheme        = localSchemeBuilder.AddToScheme
)

func init() {
	// We only register manually written functions here. The registration of the
	// generated functions takes place in the generated files. The separation
	// makes the code compile even when the generated files are missing.
	localSchemeBuilder.Register(addKnownTypes, addConversionFuncs)
	localSchemeBuilder.Register(addKnownTypes)
}

// Adds the list of known types to api.Scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&Instance{},
		&InstanceList{},
		&WingJob{},
		&WingJobList{},
		&PuppetTarget{},
		&PuppetTargetList{},
	)
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)

	return nil
}

func addConversionFuncs(scheme *runtime.Scheme) error {
	// Add field conversion funcs.
	err := scheme.AddFieldLabelConversionFunc("wing.tarmak.io/v1alpha1", "WingJob",
		func(label, value string) (string, string, error) {
			switch label {
			case "metadata.name",
				"metadata.namespace",
				"spec.instanceName":
				return label, value, nil
			default:
				return "", "", fmt.Errorf("field label not supported: %s", label)
			}
		},
	)
	if err != nil {
		return err
	}
	return nil
}

// Resource takes an unqualified resource and returns a Group qualified GroupResource
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}
