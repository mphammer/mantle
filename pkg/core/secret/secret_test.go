package secret

import (
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// map[SecretName]V1SecretName
var mappings = map[string]string{
	"Name":        "Name",
	"Namespace":   "Namespace",
	"Version":     "APIVersion",
	"Cluster":     "ClusterName",
	"Labels":      "Labels",
	"Annotations": "Annotations",
	"Data":        "Data",
	"StringData":  "StringData",
	"SecretType":  "Type",
}

// TestNewSecretFromKubeSecret verifies that NewSecretFromKubeSecret()
// successfully creates a Secret from a kubernetes Secret
func TestNewSecretFromKubeSecret(t *testing.T) {
	testcases := []struct {
		description string
		obj         interface{}
	}{
		{
			description: "v1 secret object",
			obj:         v1.Secret{},
		},
		{
			description: "v1 secret pointer",
			obj:         &v1.Secret{},
		},
	}

	for _, tc := range testcases {
		obj, _ := NewSecretFromKubeSecret(tc.obj)
		expectedObj := reflect.TypeOf(&Secret{})
		objType := reflect.TypeOf(obj)
		if expectedObj != objType {
			t.Errorf("expected %s got %s", expectedObj, objType)
		}
	}
}

// TestFromKubeSecretV1 verifies that fromKubeSecretV1() correctly populates
// the v1.Secret{} fields
func TestFromKubeSecretV1(t *testing.T) {
	v1S := v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        "testS",
			Namespace:   "testNS",
			ClusterName: "testCluster",
			Labels:      map[string]string{"label1": "test1", "label2": "test2"},
			Annotations: map[string]string{"ann1": "test1", "ann2": "test2"},
		},
		StringData: map[string]string{"field1": "data1", "field2": "data2"},
		Data:       map[string][]byte{"bfield1": []byte("bdata1")},
		Type:       v1.SecretTypeOpaque,
	}

	s, _ := fromKubeSecretV1(&v1S)
	if s.Name != v1S.Name {
		t.Errorf("incorrect name, expected %s got %s", v1S.Name, s.Name)
	}

	for name, v1Name := range mappings {
		value := reflect.ValueOf(s).Elem().FieldByName(name).Interface()
		v1Value := reflect.ValueOf(v1S).FieldByName(v1Name).Interface()
		if !reflect.DeepEqual(value, v1Value) {
			t.Errorf("incorrect %s, expected %s, got %s", name, v1Value, value)
		}
	}
}

// TestToKube verifies that ToKube() successfully returns
// a v1.Secret{}
func TestToKube(t *testing.T) {
	testcases := []struct {
		description string
		version     string
		expectedObj interface{}
	}{
		{
			description: "v1 api version",
			version:     "v1",
			expectedObj: &v1.Secret{},
		},
		{
			description: "empty api version",
			version:     "",
			expectedObj: &v1.Secret{},
		},
		{
			description: "unknown api version",
			version:     "unknown",
			expectedObj: nil,
		},
	}

	for _, tc := range testcases {
		s := Secret{
			Version: tc.version,
		}
		kubeObj, err := s.ToKube()
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

// TestToKubeV1 verifies that toKubev1() correctly populates
// the v1.Secret{} fields
func TestToKubeV1(t *testing.T) {
	s := Secret{
		Version:     "v1",
		Name:        "testS",
		Namespace:   "testNS",
		Cluster:     "testCluster",
		Labels:      map[string]string{"label1": "test1", "label2": "test2"},
		Annotations: map[string]string{"ann1": "test1", "ann2": "test2"},
		StringData:  map[string]string{"field1": "data1", "field2": "data2"},
		Data:        map[string][]byte{"bfield1": []byte("bdata1")},
		SecretType:  v1.SecretTypeOpaque,
	}

	kubeObj, _ := s.toKubeV1()
	for name, v1Name := range mappings {
		value := reflect.ValueOf(s).FieldByName(name).Interface()
		v1Value := reflect.ValueOf(kubeObj).Elem().FieldByName(v1Name).Interface()
		if !reflect.DeepEqual(value, v1Value) {
			t.Errorf("incorrect %s, expected %s, got %s", v1Name, value, v1Value)
		}
	}
}

var mapToSecretType = map[v1.SecretType]SecretType{
	v1.SecretTypeOpaque:              SecretTypeOpaque,
	v1.SecretTypeServiceAccountToken: SecretTypeServiceAccountToken,
	v1.SecretTypeDockercfg:           SecretTypeDockercfg,
	v1.SecretTypeDockerConfigJson:    SecretTypeDockerConfigJson,
	v1.SecretTypeBasicAuth:           SecretTypeBasicAuth,
	v1.SecretTypeSSHAuth:             SecretTypeSSHAuth,
	v1.SecretTypeTLS:                 SecretTypeTLS,
}

// TestConvertSecretType verifies converting to Secret Types from Kubernetes
func TestConvertSecretType(t *testing.T) {
	for kubeSecretType, secretType := range mapToSecretType {
		gotType := convertSecretType(kubeSecretType)
		if gotType != secretType {
			t.Errorf("incorrect conversion of %s, expected %s, got %s", kubeSecretType, secretType, gotType)
		}
	}
}

var mapToKubeSecretType = map[SecretType]v1.SecretType{
	SecretTypeOpaque:              v1.SecretTypeOpaque,
	SecretTypeServiceAccountToken: v1.SecretTypeServiceAccountToken,
	SecretTypeDockercfg:           v1.SecretTypeDockercfg,
	SecretTypeDockerConfigJson:    v1.SecretTypeDockerConfigJson,
	SecretTypeBasicAuth:           v1.SecretTypeBasicAuth,
	SecretTypeSSHAuth:             v1.SecretTypeSSHAuth,
	SecretTypeTLS:                 v1.SecretTypeTLS,
}

// TestRevertSecretType verifies converting to Kubernetes Secret Types
func TestRevertSecretType(t *testing.T) {
	for secretType, KubeSecretType := range mapToKubeSecretType {
		gotType := revertSecretType(secretType)
		if gotType != secretType {
			t.Errorf("incorrect conversion of %s, expected %s, got %s", secretType, KubeSecretType, gotType)
		}
	}
}
