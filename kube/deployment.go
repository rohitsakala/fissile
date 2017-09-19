package kube

import (
	"github.com/SUSE/fissile/helm"
	"github.com/SUSE/fissile/model"
)

// NewDeployment creates a Deployment for the given role, and its attached services
func NewDeployment(role *model.Role, settings *ExportSettings) (helm.Node, helm.Node, error) {
	podTemplate, err := NewPodTemplate(role, settings)
	if err != nil {
		return nil, nil, err
	}

	svc, err := NewClusterIPServiceList(role, false, settings)
	if err != nil {
		return nil, nil, err
	}

	spec := helm.NewEmptyMapping()
	spec.AddInt("replicas", role.Run.Scaling.Min)
	spec.AddNode("selector", newSelector(role.Name))
	spec.AddNode("template", podTemplate)

	deployment := newKubeConfig("extensions/v1beta1", "Deployment", role.Name)
	deployment.AddNode("spec", spec)

	return deployment.Sort(), svc, nil
}
