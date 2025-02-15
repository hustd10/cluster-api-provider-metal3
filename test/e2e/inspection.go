package e2e

import (
	"context"

	bmov1alpha1 "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/cluster-api/test/framework"
	"sigs.k8s.io/cluster-api/test/framework/clusterctl"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	inspectAnnotation = "inspect.metal3.io"
)

type InspectionInput struct {
	E2EConfig             *clusterctl.E2EConfig
	ClusterctlConfigPath  string
	BootstrapClusterProxy framework.ClusterProxy
	Namespace             string
	SpecName              string
}

func inspection(ctx context.Context, inputGetter func() InspectionInput) {
	Logf("Starting inspection tests")
	input := inputGetter()
	var (
		numberOfWorkers       = int(*input.E2EConfig.GetInt32PtrVariable("WORKER_MACHINE_COUNT"))
		numberOfAvailableBMHs = 2 * numberOfWorkers
	)

	bootstrapClient := input.BootstrapClusterProxy.GetClient()

	Logf("Request inspection for all Available BMHs via API")
	availableBMHList := bmov1alpha1.BareMetalHostList{}
	Expect(bootstrapClient.List(ctx, &availableBMHList, client.InNamespace(input.Namespace))).To(Succeed())
	Logf("Request inspection for all Available BMHs via API")
	for _, bmh := range availableBMHList.Items {
		if bmh.Status.Provisioning.State == bmov1alpha1.StateAvailable {
			AnnotateBmh(ctx, bootstrapClient, bmh, inspectAnnotation, pointer.String(""))
		}
	}

	WaitForNumBmhInState(ctx, bmov1alpha1.StateInspecting, WaitForNumInput{
		Client:    bootstrapClient,
		Options:   []client.ListOption{client.InNamespace(input.Namespace)},
		Replicas:  numberOfAvailableBMHs,
		Intervals: input.E2EConfig.GetIntervals(input.SpecName, "wait-bmh-inspecting"),
	})

	WaitForNumBmhInState(ctx, bmov1alpha1.StateAvailable, WaitForNumInput{
		Client:    bootstrapClient,
		Options:   []client.ListOption{client.InNamespace(input.Namespace)},
		Replicas:  numberOfAvailableBMHs,
		Intervals: input.E2EConfig.GetIntervals(input.SpecName, "wait-bmh-available"),
	})

	By("INSPECTION TESTS PASSED!")
}
