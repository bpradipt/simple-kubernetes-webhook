package mutation

import (
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/slackhq/simple-kubernetes-webhook/pkg/utils"
	corev1 "k8s.io/api/core/v1"
)

const (
	VM_DEFAULT_CPU_MILLIVALUE = 1000
	VM_DEFAULT_MEM            = 16777216
)

// editResourceSpec is a container for the mutation modifying resource spec
type editResourceSpec struct {
	Logger logrus.FieldLogger
}

// editResourceSpec implements the podMutator interface
var _ podMutator = (*editResourceSpec)(nil)

// Name returns the struct name
func (se editResourceSpec) Name() string {
	return "edit_resourcespec"
}

// Mutate returns a new mutated pod according to set env rules
func (se editResourceSpec) Mutate(pod *corev1.Pod) (*corev1.Pod, error) {
	se.Logger = se.Logger.WithField("mutation", se.Name())
	mpod := pod.DeepCopy()
	//var vmCpuTotal, vmMemTotal resource.Quantity

	vmCpuTotal := utils.GetResourceRequest(mpod, corev1.ResourceCPU)
	vmMemTotal := utils.GetResourceRequest(mpod, corev1.ResourceMemory)

	if vmCpuTotal < VM_DEFAULT_CPU_MILLIVALUE {
		vmCpuTotal = VM_DEFAULT_CPU_MILLIVALUE
	}
	if vmMemTotal < VM_DEFAULT_MEM {
		vmMemTotal = VM_DEFAULT_MEM
	}

	/*
		for _, container := range mpod.Spec.Containers {
			se.Logger.Debugf("container details %s", container)
			if container.Resources.Limits != nil {

				vmCpuTotal.Add(container.Resources.Limits[corev1.ResourceCPU])
				vmMemTotal.Add(container.Resources.Limits[corev1.ResourceMemory])
			}

			if container.Resources.Requests != nil && container.Resources.Limits == nil {
				vmCpuTotal.Add(container.Resources.Requests[corev1.ResourceCPU])
				vmMemTotal.Add(container.Resources.Requests[corev1.ResourceMemory])
			}

			if container.Resources.Requests == nil && container.Resources.Limits == nil {
				// Set default VM CPU and Mem resources
				vmCpuTotal, _ = resource.ParseQuantity(VM_DEFAULT_CPU)
				vmMemTotal, _ = resource.ParseQuantity(VM_DEFAULT_MEM)
			}
			//Remove the resource spec details
			removeResourceSpec(container)
		}
	*/
	// Add vmMemTotal and vmCpuTotal as annotation to the POD

	if mpod.Annotations == nil {
		mpod.Annotations = map[string]string{}
	}

	mpod.Annotations["kata.peerpods.io/vmcpu"] = strconv.FormatInt(vmCpuTotal, 10)
	mpod.Annotations["kata.peerpods.io/vmmem"] = strconv.FormatInt(vmMemTotal, 10)

	for idx, _ := range mpod.Spec.Containers {
		mpod.Spec.Containers[idx].Resources = corev1.ResourceRequirements{}
		//removeResourceSpec(&container)
	}
	se.Logger.Debugf("Updated POD details %v", mpod)
	return mpod, nil
}

// editResourceSpec injects a var in both containers and init containers of a pod
func removeResourceSpec(container *corev1.Container) {
	container.Resources = corev1.ResourceRequirements{}
}

// HasResourceSpec returns true if environment variable exists false otherwise
func HasResourceSpec(container corev1.Container) bool {
	if container.Resources.Requests == nil && container.Resources.Limits == nil {
		return false
	}
	return true
}
