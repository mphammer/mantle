package namespace

import (
	"fmt"
	"reflect"

	serrors "github.com/koki/structurederrors"
	"k8s.io/api/core/v1"
)

// NewNamespaceFromKubeNamespace will create a new Namespace object with
// the data from a provided kubernetes namespace object
func NewNamespaceFromKubeNamespace(ns interface{}) (*Namespace, error) {
	switch reflect.TypeOf(ns) {
	case reflect.TypeOf(v1.Namespace{}):
		obj := ns.(v1.Namespace)
		if obj.APIVersion != "v1" {
			return nil, fmt.Errorf("mis-matched versions.  Namespace type: %s, APIVersion: %s", reflect.TypeOf(ns), obj.APIVersion)
		}
		return fromKubeNamespaceV1(&obj)
	case reflect.TypeOf(&v1.Namespace{}):
		obj := ns.(*v1.Namespace)
		if obj.APIVersion != "v1" {
			return nil, fmt.Errorf("mis-matched versions.  Namespace type: %s, APIVersion: %s", reflect.TypeOf(ns), obj.APIVersion)
		}
		return fromKubeNamespaceV1(obj)
	default:
		return nil, fmt.Errorf("unknown Namespace version: %s", reflect.TypeOf(ns))
	}
}

// fromKubeNamespaceV1 converts to koki namespace for V1
func fromKubeNamespaceV1(kubeNamespace *v1.Namespace) (*Namespace, error) {
	kokiNamespace := Namespace{}

	kokiNamespace.Name = kubeNamespace.Name
	kokiNamespace.Namespace = kubeNamespace.Namespace
	kokiNamespace.Version = kubeNamespace.APIVersion
	kokiNamespace.Cluster = kubeNamespace.ClusterName
	kokiNamespace.Labels = kubeNamespace.Labels
	kokiNamespace.Annotations = kubeNamespace.Annotations

	finalizers, err := fromKubeNamespaceSpecV1(kubeNamespace.Spec)
	if err != nil {
		return nil, err
	}
	kokiNamespace.Finalizers = finalizers

	phase, err := fromKubeNamespaceStatusV1(kubeNamespace.Status)
	if err != nil {
		return nil, err
	}
	kokiNamespace.Phase = phase

	return &kokiNamespace, nil
}

// fromKubeNamespaceSpecV1 changes a kubernetes spec to a kubernetes finalizer
func fromKubeNamespaceSpecV1(kubeSpec v1.NamespaceSpec) ([]FinalizerName, error) {
	var kokiFinalizers []FinalizerName

	for i := range kubeSpec.Finalizers {
		kubeFinalizer := kubeSpec.Finalizers[i]

		var kokiFinalizer FinalizerName
		switch kubeFinalizer {
		case v1.FinalizerKubernetes:
			kokiFinalizer = FinalizerKubernetes
		default:
			return nil, serrors.InvalidValueErrorf(kubeFinalizer, "unrecognized finalizer")
		}

		kokiFinalizers = append(kokiFinalizers, kokiFinalizer)
	}

	return kokiFinalizers, nil
}

// fromKubeNamespaceStatusV1 changes a kubernetes status to a kubernetes phase
func fromKubeNamespaceStatusV1(kubeStatus v1.NamespaceStatus) (NamespacePhase, error) {
	switch kubeStatus.Phase {
	case v1.NamespaceActive:
		return NamespaceActive, nil
	case v1.NamespaceTerminating:
		return NamespaceTerminating, nil
	}
	return NamespaceActive, serrors.InvalidValueErrorf(kubeStatus.Phase, "invalid phase")
}
