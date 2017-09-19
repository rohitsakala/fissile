package kube

import (
	"fmt"

	"github.com/SUSE/fissile/helm"
	"github.com/SUSE/fissile/model"
)

const volumeStorageClassAnnotation = "volume.beta.kubernetes.io/storage-class"

// NewStatefulSet returns a stateful set for the given role
func NewStatefulSet(role *model.Role, settings *ExportSettings) (helm.Node, helm.Node, error) {
	// For each StatefulSet, we need two services -- one for the public (inside
	// the namespace) endpoint, and one headless service to control the pods.
	if role == nil {
		panic(fmt.Sprintf("No role given"))
	}

	podTemplate, err := NewPodTemplate(role, settings)
	if err != nil {
		return nil, nil, err
	}

	svcList, err := NewClusterIPServiceList(role, true, settings)
	if err != nil {
		return nil, nil, err
	}

	claims := getAllVolumeClaims(role, settings.CreateHelmChart)

	spec := helm.NewIntMapping("replicas", role.Run.Scaling.Min)
	spec.Add("serviceName", fmt.Sprintf("%s-set", role.Name))
	spec.AddNode("template", podTemplate)
	spec.AddNode("volumeClaimTemplates", helm.NewNodeList(claims...))

	statefulSet := newKubeConfig("apps/v1beta1", "StatefulSet", role.Name)
	statefulSet.AddNode("spec", spec)

	return statefulSet.Sort(), svcList, nil
}

// getAllVolumeClaims returns the list of persistent and shared volume claims from a role
func getAllVolumeClaims(role *model.Role, createHelmChart bool) []helm.Node {
	claims := getVolumeClaims(role.Run.PersistentVolumes, "persistent", "ReadWriteOnce", createHelmChart)
	claims = append(claims, getVolumeClaims(role.Run.SharedVolumes, "shared", "ReadWriteMany", createHelmChart)...)
	return claims
}

// getVolumeClaims returns the list of persistent volume claims from a role
func getVolumeClaims(volumeDefinitions []*model.RoleRunVolume, storageClass string, accessMode string, createHealmChart bool) []helm.Node {
	if createHealmChart {
		storageClass = fmt.Sprintf("{{ .Values.kube.storage_class.%s | quote }}", storageClass)
	}

	var claims []helm.Node
	for _, volume := range volumeDefinitions {
		meta := helm.NewMapping("name", volume.Tag)
		meta.AddNode("annotations", helm.NewMapping(volumeStorageClassAnnotation, storageClass))

		size := fmt.Sprintf("%dG", volume.Size)

		spec := helm.NewNodeMapping("accessModes", helm.NewList(accessMode))
		spec.AddNode("resources", helm.NewNodeMapping("requests", helm.NewMapping("storage", size)))

		claim := helm.NewNodeMapping("metadata", meta)
		claim.AddNode("spec", spec)

		claims = append(claims, claim)
	}
	return claims
}
