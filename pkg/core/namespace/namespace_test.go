package namespace

import (
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var metaDataMappings = map[string]string{
	"Name":        "Name",
	"Namespace":   "Namespace",
	"Version":     "APIVersion",
	"Cluster":     "ClusterName",
	"Labels":      "Labels",
	"Annotations": "Annotations",
}

// TestNewNamespaceFromKubeNamespace verifies that the correct version and type are created
func TestNewNamespaceFromKubeNamespace(t *testing.T) {
	testcases := []struct {
		description string
		obj         interface{}
	}{
		{
			description: "v1 namespace object",
			obj:         v1.Namespace{},
		},
		{
			description: "v1 namespace pointer",
			obj:         &v1.Namespace{},
		},
	}

	for _, tc := range testcases {
		obj, _ := NewNamespaceFromKubeNamespace(tc.obj)
		expectedObj := reflect.TypeOf(&Namespace{})
		objType := reflect.TypeOf(obj)
		if expectedObj != objType {
			t.Errorf("incorrect koki namespace, expected %s, got %s", expectedObj, objType)
		}
	}
}

// TestFromKubeV1 verifies that the fields were correctly populated
func TestFromKubeV1(t *testing.T) {
	v1Ns := v1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        "testCM",
			Namespace:   "testNS",
			ClusterName: "testCluster",
			Labels:      map[string]string{"label1": "test1", "label2": "test2"},
			Annotations: map[string]string{"ann1": "test1", "ann2": "test2"},
		},
		Spec: v1.NamespaceSpec{
			Finalizers: []v1.FinalizerName{
				v1.FinalizerKubernetes,
			},
		},
		Status: v1.NamespaceStatus{
			Phase: v1.NamespaceActive,
		},
	}

	ns, _ := fromKubeNamespaceV1(&v1Ns)

	// Check Meta Data Fields
	for name, v1Name := range metaDataMappings {
		value := reflect.ValueOf(ns).Elem().FieldByName(name).Interface()
		v1Value := reflect.ValueOf(v1Ns).FieldByName(v1Name).Interface()
		if !reflect.DeepEqual(value, v1Value) {
			t.Errorf("incorrect koki field %s, expected %s, got %s", name, v1Value, value)
		}
	}
	// Check Finalizers
	valueFinalizer := ns.Finalizers 
	v1ValueFinalizer := v1Ns.Spec.Finalizers
	if !finalizersEqual(&valueFinalizer, &v1ValueFinalizer) {
		t.Errorf("incorrect koki finalizers, used %+v, got %+v", v1ValueFinalizer, valueFinalizer)
	}
	// Check NamespacePhase
	valuePhase := ns.Phase 
	if valuePhase != NamespaceActive {
		t.Errorf("incorrect koki Phase, expectec %+v, got %+v", NamespaceActive, valuePhase)
	}
}

// TestToKube verifies that the correct version and type were returned
func TestToKube(t *testing.T) {
	testcases := []struct {
		description string
		version     string
		expectedObj interface{}
	}{
		{
			description: "v1 api version",
			version:     "v1",
			expectedObj: &v1.Namespace{},
		},
		{
			description: "empty api version",
			version:     "",
			expectedObj: &v1.Namespace{},
		},
		{
			description: "unknown api version",
			version:     "unknown",
			expectedObj: nil,
		},
	}

	for _, tc := range testcases {
		ns := Namespace{
			Version: tc.version,
		}
		kubeObj, err := ns.ToKube()
		kubeType := reflect.TypeOf(kubeObj)
		expectedType := reflect.TypeOf(tc.expectedObj)
		if kubeType != expectedType {
			t.Errorf("wrong api version, got %s expected %s", kubeType, expectedType)
		}
		if tc.expectedObj == nil && err == nil {
			t.Errorf("no error returned")
		}
	}
}

// TestToKubeV1 verifies that the fields were correctly populated
func TestToKubeV1(t *testing.T) {
	ns := Namespace{
		Version:     "v1",
		Name:        "testNS",
		Namespace:   "testNS",
		Cluster:     "testCluster",
		Labels:      map[string]string{"label1": "test1", "label2": "test2"},
		Annotations: map[string]string{"ann1": "test1", "ann2": "test2"},
		Finalizers:  []FinalizerName{FinalizerKubernetes},
		Phase:       NamespaceTerminating,
	}

	kubeObj, _ := ns.toKubeV1()
	
	// Check Meta Data Fields
	for name, v1Name := range metaDataMappings {
		value := reflect.ValueOf(ns).FieldByName(name).Interface()
		v1Value := reflect.ValueOf(kubeObj).Elem().FieldByName(v1Name).Interface()
		if !reflect.DeepEqual(value, v1Value) {
			t.Errorf("incorrect %s, expected %s, got %s", v1Name, value, v1Value)
		}
	}
	// Check Finalizers
	valueFinalizer := ns.Finalizers 
	v1ValueFinalizer := kubeObj.Spec.Finalizers
	if !finalizersEqual(&valueFinalizer, &v1ValueFinalizer) {
		t.Errorf("incorrect converting to kube finalizers, used %+v, got %+v", valueFinalizer, v1ValueFinalizer)
	}
	// Check NamespacePhase
	v1ValuePhase := kubeObj.Status.Phase
	if v1ValuePhase != v1.NamespaceTerminating {
		t.Errorf("incorrect converting to kube Phase, expected %+v, got %+v", v1.NamespaceTerminating, v1ValuePhase)
	}
}

// TestFromKubeNamespaceSpecV1 verifies the correct type and values were returned
func TestFromKubeNamespaceSpecV1(t *testing.T) {
	kubeSpec1 := v1.NamespaceSpec{Finalizers: []v1.FinalizerName{v1.FinalizerKubernetes}}
	kubeSpec2 := v1.NamespaceSpec{Finalizers: []v1.FinalizerName{}}
	kubeSpecMap := map[*v1.NamespaceSpec][]FinalizerName{
		&kubeSpec1: []FinalizerName{FinalizerKubernetes},
		&kubeSpec2: nil,
	}
	for kubeSpec, kokiFinalizers := range kubeSpecMap {
		gotFinalizers, err := fromKubeNamespaceSpecV1(*kubeSpec)
		if err != nil {
			t.Errorf("Error from v1 spec: %s", err)
		}
		// Check the Type Received
		expectedType := reflect.TypeOf([]FinalizerName{})
		gotType := reflect.TypeOf(gotFinalizers)
		if expectedType != gotType {
			t.Errorf("wrong koki finalizers type, got %s expected %s", gotType, expectedType)
		}
		// Check the Value Received
		if !reflect.DeepEqual(gotFinalizers, kokiFinalizers) {
			t.Errorf("wrong finalizer values, expected %+v, got %+v", kokiFinalizers, gotFinalizers)
		}
	}

}

// TestFromKubeNamespaceStatusV1 verifies the correct type and values were returned
func TestFromKubeNamespaceStatusV1(t *testing.T) {
	kubeStatus1 := v1.NamespaceStatus{Phase: v1.NamespaceActive}
	kubeStatus2 := v1.NamespaceStatus{Phase: v1.NamespaceTerminating}
	kubeStatusMap := map[*v1.NamespaceStatus]NamespacePhase{
		&kubeStatus1: NamespaceActive,
		&kubeStatus2: NamespaceTerminating,
	}
	for kubeStatus, kokiPhase := range kubeStatusMap {
		gotKokiPhase, err := fromKubeNamespaceStatusV1(*kubeStatus)
		if err != nil {
			t.Errorf("Error from v1 status: %s", err)
		}
		// Check the Type Received
		expectedType := reflect.TypeOf(NamespaceActive)
		gotType := reflect.TypeOf(gotKokiPhase)
		if expectedType != gotType {
			t.Errorf("wrong koki Phase type, got %+v expected %+v", gotType, expectedType)
		}
		// Check the Value Received
		if gotKokiPhase != kokiPhase {
			t.Errorf("wrong phase value, expected %+v, got %+v", kokiPhase, gotKokiPhase)
		}
	}
}
 // TestToKubeNamespaceStatusV1 verifies the correct type and values were returned
func TestToKubeNamespaceStatusV1(t *testing.T) {
	kokiNamespaces := map[*Namespace]v1.NamespacePhase{
		&Namespace{Phase: NamespaceActive}:      v1.NamespaceActive,
		&Namespace{Phase: NamespaceTerminating}: v1.NamespaceTerminating,
	}
	for kokiNamespace, kubePhase := range kokiNamespaces {
		gotKubeStatus, err := kokiNamespace.toKubeNamespaceStatusV1()
		if err != nil {
			t.Errorf("Error to v1 status: %s", err)
		}
		// Check the Type Received
		expectedType := reflect.TypeOf(v1.NamespaceStatus{})
		gotType := reflect.TypeOf(gotKubeStatus)
		if expectedType != gotType {
			t.Errorf("wrong kube NamespaceStatus type, got %s expected %s", gotType, expectedType)
		}
		// Check the Value Received
		if string(gotKubeStatus.Phase) != string(kubePhase) {
			t.Errorf("wrong finalizers in spec")
		}
	}
}

// TestToKubeNamespaceSpecV1 verifies the correct type and values were returned
func TestToKubeNamespaceSpecV1(t *testing.T) {
	kokiNamespaces := []Namespace{
		Namespace{Finalizers: []FinalizerName{}},                    // no finalizers
		Namespace{Finalizers: []FinalizerName{FinalizerKubernetes}}, // one finalizer
	}
	for _, kokiNamespace := range kokiNamespaces {
		gotKubeNamespaceSpec, err := kokiNamespace.toKubeNamespaceSpecV1()
		if err != nil {
			t.Errorf("Error to v1 spec: %s", err)
		}
		// Check the Type Received
		expectedType := reflect.TypeOf(v1.NamespaceSpec{})
		gotType := reflect.TypeOf(gotKubeNamespaceSpec)
		if expectedType != gotType {
			t.Errorf("wrong kube NamespaceSpec type, got %s expected %s", gotType, expectedType)
		}
		// Check the Values Received
		if !finalizersEqual(&kokiNamespace.Finalizers, &gotKubeNamespaceSpec.Finalizers) {
			t.Errorf("wrong finalizers in spec")
		}
	}
}

// TestToKubeFinalizerV1 
func TestToKubeFinalizerV1(t *testing.T) {
	finalizers := map[FinalizerName]v1.FinalizerName{
		FinalizerKubernetes: v1.FinalizerKubernetes,
	}
	for kokiFinalizer, kubeFinalizer := range finalizers {
		gotKubeFinalizer, err := toKubeFinalizerV1(kokiFinalizer)
		if err != nil {
			t.Errorf("%Error to v1 finalizer: s", err)
		}
		// Check the Type Received
		expectedType := reflect.TypeOf(kubeFinalizer)
		gotType := reflect.TypeOf(gotKubeFinalizer)
		if expectedType != gotType {
			t.Errorf("wrong kube finalizer type, got %s, expected %s", gotType, expectedType)
		}
		// Check the Value Received
		if !reflect.DeepEqual(gotKubeFinalizer, kubeFinalizer) {
			t.Errorf("incorrect finalizer, expected %s, got %s", kubeFinalizer, gotKubeFinalizer)
		}
	}
}

// TestFinalizersEqual verifies a list of koki finalizers can be compared to a list of kube finalizers
func TestFinalizersEqual(t *testing.T) {
	kokiFinalizers1 := []FinalizerName{}
	kokiFinalizers2 := []FinalizerName{FinalizerKubernetes}
	kubeFinalizers1 := []v1.FinalizerName{}
	kubeFinalizers2 := []v1.FinalizerName{v1.FinalizerKubernetes}
	if !finalizersEqual(&kokiFinalizers1, &kubeFinalizers1) {
		t.Errorf("incorrect finalizer equality for empty, expected %+v, got %+v", true, false)
	}
	if !finalizersEqual(&kokiFinalizers2, &kubeFinalizers2) {
		t.Errorf("incorrect finalizer equality for 1 finalizer each, expected %+v, got %+v", true, false)
	}
	if finalizersEqual(&kokiFinalizers1, &kubeFinalizers2) {
		t.Errorf("incorrect finalizer equality for different number of finalizers, expected %+v, got %+v", false, true)
	}
}


// finalizersEqual checks that lists of finalizers are equal, handles if out of order
func finalizersEqual(kokiFinalizers *[]FinalizerName, kubeFinalizers *[]v1.FinalizerName) bool {
	if ((kokiFinalizers == nil) != (kubeFinalizers == nil)) || (len(*kokiFinalizers) != len(*kubeFinalizers)) {
		return false
	}
	kubeFinalizersMap := make([]bool, len(*kubeFinalizers), len(*kubeFinalizers))
	for _, kokiFinalizer := range *kokiFinalizers {
		found := false
		for i, kubeFinalizer := range *kubeFinalizers {
			convertedKokiFinalizer, _ := toKubeFinalizerV1(kokiFinalizer)
			if reflect.DeepEqual(convertedKokiFinalizer, kubeFinalizer) && !kubeFinalizersMap[i] {
				kubeFinalizersMap[i] = true // mark as found
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
