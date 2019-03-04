package secret

import (
	"fmt"
	"strings"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	serrors "github.com/koki/structurederrors"
)

// ToKube will return a kubernetes secret object of the
// api version type defined in the object
func (s *Secret) ToKube() (runtime.Object, error) {
	switch strings.ToLower(s.Version) {
	case "v1":
		return s.toKubeV1()
	case "":
		return s.toKubeV1()
	default:
		return nil, fmt.Errorf("unsupported api version for secret: %s", cm.Version)
	}
}

func (s *Secret) toKubeV1() (*v1.Secret, error) {
	kubeSecret := &v1.Secret{}
	
	kubeSecret.Name = s.Name
	kubeSecret.Namespace = s.Namespace
	kubeSecret.APIVersion = s.Version
	kubeSecret.ClusterName = s.Cluster
	kubeSecret.Kind = "Secret"
	kubeSecret.Labels = s.Labels
	kubeSecret.Annotations = s.Annotations
	kubeSecret.Data = s.Data 
	kubeSecret.StringData = s.StringData
	type, err := revertSecretType(s.SecretType)
	if err != nil {
		return nil, err
	}
	kubeSecret.Type = type

	return kubeSecret, nil
}

// revertSecretType converts to a kubernetes SecretType
func revertSecretType(secret SecretType) (v1.SecretType, error) {
	if secret == "" {
		return "", nil
	}

	switch secret {
	case SecretTypeOpaque:
		return v1.SecretTypeOpaque, nil
	case SecretTypeServiceAccountToken:
		return v1.SecretTypeServiceAccountToken, nil
	case SecretTypeDockercfg:
		return v1.SecretTypeDockercfg, nil
	case SecretTypeDockerConfigJson:
		return v1.SecretTypeDockerConfigJson, nil
	case SecretTypeBasicAuth:
		return v1.SecretTypeBasicAuth, nil
	case SecretTypeSSHAuth:
		return v1.SecretTypeSSHAuth, nil
	case SecretTypeTLS:
		return v1.SecretTypeTLS, nil
	default:
		return "", serrors.InvalidValueErrorf(secret, "unrecognized Secret type")
	}
}
