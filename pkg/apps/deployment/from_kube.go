package deployment

import (
	v1 "k8s.io/api/apps/v1"
	v1beta1 "k8s.io/api/apps/v1beta1"
	v1beta2 "k8s.io/api/apps/v1beta2"
	
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/koki/short/parser"
	"github.com/koki/short/types"
	"github.com/koki/short/yaml"
	serrors "github.com/koki/structurederrors"
)

// NewDeploymentFromKubeDeployment will create a new Deployment object with
// the data from a provided kubernetes Deployment object
func NewDeploymentFromKubeDeployment(d interface{}) (*Deployment, error) {
	switch reflect.TypeOf(d) {
	case reflect.TypeOf(v1.Deployment{}):
		obj := d.(v1.Deployment)
		if obj.APIVersion != "v1" {
			return nil, fmt.Errorf("mis-matched versions.  Deployment type: %s, APIVersion: %s", reflect.TypeOf(d), obj.APIVersion)
		}
		return fromKubeDeploymentV1(&obj)
	case reflect.TypeOf(&v1.Deployment{}):
		obj := d.(*v1.Deployment)
		if obj.APIVersion != "v1" {
			return nil, fmt.Errorf("mis-matched versions.  Deployment type: %s, APIVersion: %s", reflect.TypeOf(d), obj.APIVersion)
		}
		return fromKubeDeploymentV1(obj)
	case reflect.TypeOf(v1beta1.Deployment{}):
		obj := d.(v1beta1.Deployment)
		if obj.APIVersion != "v1beta1" {
			return nil, fmt.Errorf("mis-matched versions.  Deployment type: %s, APIVersion: %s", reflect.TypeOf(d), obj.APIVersion)
		}
		return fromKubeDeploymentV1Beta1(&obj)
	case reflect.TypeOf(&v1beta1.Deployment{}):
		obj := d.(*v1beta1.Deployment)
		if obj.APIVersion != "v1beta1" {
			return nil, fmt.Errorf("mis-matched versions.  Deployment type: %s, APIVersion: %s", reflect.TypeOf(d), obj.APIVersion)
		}
		return fromKubeDeploymentV1Beta1(obj)
	case reflect.TypeOf(v1beta2.Deployment{}):
		obj := d.(v1beta2.Deployment)
		if obj.APIVersion != "v1beta2" {
			return nil, fmt.Errorf("mis-matched versions.  Deployment type: %s, APIVersion: %s", reflect.TypeOf(d), obj.APIVersion)
		}
		return fromKubeDeploymentV1Beta2(&obj)
	case reflect.TypeOf(&v1beta2.Deployment{}):
		obj := d.(*v1beta2.Deployment)
		if obj.APIVersion != "v1beta2" {
			return nil, fmt.Errorf("mis-matched versions.  Deployment type: %s, APIVersion: %s", reflect.TypeOf(d), obj.APIVersion)
		}
		return fromKubeDeploymentV1Beta2(obj)
	default:
		return nil, fmt.Errorf("unknown Deployment version: %s", reflect.TypeOf(d))
	}
}

func fromKubeDeploymentV1(kubeDeployment runtime.Object) (Deployment, error) {

}

func fromKubeDeploymentV1Beta1(kubeDeployment runtime.Object) (Deployment, error) {

}

func fromKubeDeploymentV1Beta2(kubeDeployment runtime.Object) (Deployment, error) {
	kokiDeployment := &Deployment{}

	kokiDeployment.Name = kubeDeployment.Name
	kokiDeployment.Namespace = kubeDeployment.Namespace
	kokiDeployment.Version = kubeDeployment.APIVersion
	kokiDeployment.Cluster = kubeDeployment.ClusterName
	kokiDeployment.Labels = kubeDeployment.Labels
	kokiDeployment.Annotations = kubeDeployment.Annotations

	kubeSpec := &kubeDeployment.Spec
	kokiDeployment.Replicas = kubeSpec.Replicas

	// Setting the Selector and Template is identical to ReplicaSet

	// Fill out the Selector and Template.Labels.
	// If kubeDeployment only has Template.Labels, we pull it up to Selector.
	selector, templateLabelsOverride, err := convertRSLabelSelector(kubeSpec.Selector, kubeSpec.Template.Labels)
	if err != nil {
		return nil, err
	}

	if selector != nil && (selector.Labels != nil || selector.Shorthand != "") {
		kokiDeployment.Selector = selector
	}

	// Build a Pod from the kube Template. Use it to set the koki Template.
	meta, template, err := convertTemplate(kubeSpec.Template)
	if err != nil {
		return nil, serrors.ContextualizeErrorf(err, "pod template")
	}
	kokiDeployment.TemplateMetadata = applyTemplateLabelsOverride(templateLabelsOverride, meta)
	kokiDeployment.PodTemplate = template

	// End Selector/Template section.

	kokiDeployment.Recreate, kokiDeployment.MaxUnavailable, kokiDeployment.MaxSurge = convertDeploymentStrategy(kubeSpec.Strategy)

	kokiDeployment.MinReadySeconds = kubeSpec.MinReadySeconds
	kokiDeployment.RevisionHistoryLimit = kubeSpec.RevisionHistoryLimit
	kokiDeployment.Paused = kubeSpec.Paused
	kokiDeployment.ProgressDeadlineSeconds = kubeSpec.ProgressDeadlineSeconds

	kokiDeployment.DeploymentStatus, err = convertDeploymentStatus(kubeDeployment.Status)
	if err != nil {
		return nil, err
	}

	return &DeploymentWrapper{
		Deployment: *kokiDeployment,
	}, nil
}

func fromDeploymentStatusV1(kubeStatus v1.DeploymentStatus) (DeploymentStatus, error) {}
func fromDeploymentStatusV1Beta1(kubeStatus v1beta2.DeploymentStatus) (DeploymentStatus, error) {}

func fromDeploymentStatusV1Beta2(kubeStatus appsv1beta2.DeploymentStatus) (DeploymentStatus, error) {
	conditions, err := convertDeploymentConditions(kubeStatus.Conditions)
	if err != nil {
		return DeploymentStatus{}, err
	}
	return DeploymentStatus{
		ObservedGeneration: kubeStatus.ObservedGeneration,
		Replicas: DeploymentReplicasStatus{
			Total:       kubeStatus.Replicas,
			Updated:     kubeStatus.UpdatedReplicas,
			Ready:       kubeStatus.ReadyReplicas,
			Available:   kubeStatus.AvailableReplicas,
			Unavailable: kubeStatus.UnavailableReplicas,
		},
		Conditions:     conditions,
		CollisionCount: kubeStatus.CollisionCount,
	}, nil
}

func fromDeploymentConditionsV1(kubeConditions []v1.DeploymentCondition) ([]DeploymentCondition, error) {}
func fromDeploymentConditionsV1Beta1(kubeConditions []v1beta2.DeploymentCondition) ([]DeploymentCondition, error) {}

func fromDeploymentConditionsV1Beta2(kubeConditions []appsv1beta2.DeploymentCondition) ([]DeploymentCondition, error) {
	if len(kubeConditions) == 0 {
		return nil, nil
	}

	kokiConditions := make([]DeploymentCondition, len(kubeConditions))
	for i, condition := range kubeConditions {
		status, err := convertConditionStatus(condition.Status)
		if err != nil {
			return nil, serrors.ContextualizeErrorf(err, "deployment conditions[%d]", i)
		}
		conditionType, err := convertDeploymentConditionType(condition.Type)
		if err != nil {
			return nil, serrors.ContextualizeErrorf(err, "deployment conditions[%d]", i)
		}
		kokiConditions[i] = DeploymentCondition{
			Type:               conditionType,
			Status:             status,
			LastUpdateTime:     condition.LastUpdateTime,
			LastTransitionTime: condition.LastTransitionTime,
			Reason:             condition.Reason,
			Message:            condition.Message,
		}
	}

	return kokiConditions, nil
}

func fromDeploymentConditionTypeV1(kubeType v1.DeploymentConditionType) (DeploymentConditionType, error) {}
func fromDeploymentConditionTypeV1Beta1(kubeType v1beta1.DeploymentConditionType) (DeploymentConditionType, error) {}

func fromDeploymentConditionTypeV1Beta2(kubeType appsv1beta2.DeploymentConditionType) (DeploymentConditionType, error) {
	switch kubeType {
	case appsv1beta2.DeploymentAvailable:
		return DeploymentAvailable, nil
	case appsv1beta2.DeploymentProgressing:
		return DeploymentProgressing, nil
	case appsv1beta2.DeploymentReplicaFailure:
		return DeploymentReplicaFailure, nil
	default:
		return DeploymentReplicaFailure, serrors.InvalidValueErrorf(kubeType, "unrecognized deployment condition type")
	}
}

func fromDeploymentStrategyV1(kubeStrategy v1.DeploymentStrategy) (isRecreate bool, maxUnavailable, maxSurge *intstr.IntOrString) {}
func fromDeploymentStrategyV1Beta1(kubeStrategy v1beta1.DeploymentStrategy) (isRecreate bool, maxUnavailable, maxSurge *intstr.IntOrString) {}

func fromDeploymentStrategyV1Beta2(kubeStrategy appsv1beta2.DeploymentStrategy) (isRecreate bool, maxUnavailable, maxSurge *intstr.IntOrString) {
	if kubeStrategy.Type == appsv1beta2.RecreateDeploymentStrategyType {
		return true, nil, nil
	}

	if rollingUpdate := kubeStrategy.RollingUpdate; rollingUpdate != nil {
		return false, rollingUpdate.MaxUnavailable, rollingUpdate.MaxSurge
	}

	return false, nil, nil
}