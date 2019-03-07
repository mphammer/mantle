package apps

import ()

type RSSelector struct {
	Shorthand string
	Labels    map[string]string
}

func fromRSLabelSelector(kubeSelector *metav1.LabelSelector, kubeTemplateLabels map[string]string) (*RSSelector, map[string]string, error) {
	// If the Selector is unspecified, it defaults to the Template's Labels.
	if kubeSelector == nil {
		return &types.RSSelector{
			Labels: kubeTemplateLabels,
		}, nil, nil
	}

	if len(kubeSelector.MatchExpressions) == 0 {
		if reflect.DeepEqual(kubeSelector.MatchLabels, kubeTemplateLabels) {
			// Selector and template labels are identical. Just keep the selector.
			return &types.RSSelector{
				Labels: kubeSelector.MatchLabels,
			}, nil, nil
		}
		return &types.RSSelector{
			Labels: kubeSelector.MatchLabels,
		}, kubeTemplateLabels, nil
	}

	selectorString, err := expressions.UnparseLabelSelector(kubeSelector)
	if err != nil {
		return nil, nil, err
	}

	return &types.RSSelector{
		Shorthand: selectorString,
	}, kubeTemplateLabels, nil
}