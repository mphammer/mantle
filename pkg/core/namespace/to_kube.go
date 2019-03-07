package namespace

import (
	"fmt"
	"strings"

	serrors "github.com/koki/structurederrors"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// ToKube will return a kubernetes namespace object of the api version
// type defined in the namespace
func (ns *Namespace) ToKube() (runtime.Object, error) {
	switch strings.ToLower(ns.Version) {
	case "v1":
		return ns.toKubeV1()
	case "":
		return ns.toKubeV1()
	default:
		return nil, fmt.Errorf("unsupported api version for Namespace: %s", ns.Version)
	}
}

// toKubev1 converts to a kubernetes namespace for V1
func (ns *Namespace) toKubeV1() (*v1.Namespace, error) {
	var err error
	kubeNamespace := &v1.Namespace{}

	kubeNamespace.Name = ns.Name
	kubeNamespace.Namespace = ns.Namespace
	if len(ns.Version) == 0 {
		kubeNamespace.APIVersion = "v1"
	} else {
		kubeNamespace.APIVersion = ns.Version
	}
	kubeNamespace.Kind = "Namespace"
	kubeNamespace.ClusterName = ns.Cluster
	kubeNamespace.Labels = ns.Labels
	kubeNamespace.Annotations = ns.Annotations

	spec, err := ns.toKubeNamespaceSpecV1()
	if err != nil {
		return nil, serrors.ContextualizeErrorf(err, "Namespace spec")
	}
	kubeNamespace.Spec = spec

	status, err := ns.toKubeNamespaceStatusV1()
	if err != nil {
		return nil, serrors.ContextualizeErrorf(err, "Namespace status")
	}
	kubeNamespace.Status = status

	return kubeNamespace, nil
}

// toKubeNamespaceStatusV1 changes a koki status to a kubernetes status
func (ns *Namespace) toKubeNamespaceStatusV1() (v1.NamespaceStatus, error) {
	var kubeStatus v1.NamespaceStatus

	if &ns.Phase == nil {
		return kubeStatus, nil
	}

	var phase v1.NamespacePhase
	switch ns.Phase {
	case NamespaceActive:
		phase = v1.NamespaceActive
	case NamespaceTerminating:
		phase = v1.NamespaceTerminating
	default:
		return kubeStatus, serrors.InvalidValueErrorf(ns.Phase, "Invalid namespace phase")
	}
	kubeStatus.Phase = phase

	return kubeStatus, nil
}

// toKubeNamespaceSpecV1 changes a koki spec to a kubernetes spec
func (ns *Namespace) toKubeNamespaceSpecV1() (v1.NamespaceSpec, error) {
	var kubeSpec v1.NamespaceSpec
	var kubeFinalizers []v1.FinalizerName

	for i := range ns.Finalizers {
		kokiFinalizer := ns.Finalizers[i]

		kubeFinalizer, err := toKubeFinalizerV1(kokiFinalizer)
		if err != nil {
			return kubeSpec, err
		}

		kubeFinalizers = append(kubeFinalizers, kubeFinalizer)
	}
	kubeSpec.Finalizers = kubeFinalizers

	return kubeSpec, nil
}

// toKubeFinalizerV1 changes a koki finalizer to a kubernetes finalizer
func toKubeFinalizerV1(kokiFinalizer FinalizerName) (v1.FinalizerName, error) {
	switch kokiFinalizer {
	case FinalizerKubernetes:
		return v1.FinalizerKubernetes, nil
	}
	return v1.FinalizerKubernetes, serrors.InvalidValueErrorf(kokiFinalizer, "unrecognized value")
}
