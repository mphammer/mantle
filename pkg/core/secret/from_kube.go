package secret

import (
	"fmt"
	"reflect"

	serrors "github.com/koki/structurederrors"
	v1 "k8s.io/api/core/v1"
)

// NewSecretFromKubeSecret will create a new Secret object with
// the data from a provided kubernetes Secret object
func NewSecretFromKubeSecret(s interface{}) (*Secret, error) {
	switch reflect.TypeOf(s) {
	case reflect.TypeOf(v1.Secret{}):
		obj := s.(v1.Secret)
		return fromKubeSecretV1(&obj)
	case reflect.TypeOf(&v1.Secret{}):
		return fromKubeSecretV1(s.(*v1.Secret))
	default:
		return nil, fmt.Errorf("unknown Secret version: %s", reflect.TypeOf(s))
	}
}

func fromKubeSecretV1(kubeSecret *v1.Secret) (*Secret, error) {
	sType, err := convertSecretType(kubeSecret.Type)
	if err != nil {
		return nil, err
	}
	s := &Secret{
		Name:        kubeSecret.Name,
		Namespace:   kubeSecret.Namespace,
		Version:     kubeSecret.APIVersion,
		Cluster:     kubeSecret.ClusterName,
		Labels:      kubeSecret.Labels,
		Annotations: kubeSecret.Annotations,
		Data:        kubeSecret.Data,
		StringData:  kubeSecret.StringData,
		SecretType:  sType,
	}

	return s, nil
}

// convertSecretType converts from a kubernetes SecretType
func convertSecretType(secret v1.SecretType) (SecretType, error) {
	if secret == "" {
		return "", nil
	}
	switch secret {
	case v1.SecretTypeOpaque:
		return SecretTypeOpaque, nil
	case v1.SecretTypeServiceAccountToken:
		return SecretTypeServiceAccountToken, nil
	case v1.SecretTypeDockercfg:
		return SecretTypeDockercfg, nil
	case v1.SecretTypeDockerConfigJson:
		return SecretTypeDockerConfigJson, nil
	case v1.SecretTypeBasicAuth:
		return SecretTypeBasicAuth, nil
	case v1.SecretTypeSSHAuth:
		return SecretTypeSSHAuth, nil
	case v1.SecretTypeTLS:
		return SecretTypeTLS, nil
	default:
		return "", serrors.InvalidValueErrorf(secret, "unrecognized Secret type")
	}
}
