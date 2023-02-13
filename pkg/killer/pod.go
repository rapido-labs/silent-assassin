package killer

import v1 "k8s.io/api/core/v1"

//filterOutPodByOwnerReferenceKind clears the pods with ownerKind specified in the second argument,
//from the slice of pods specified in the first argument.
func filterOutPodByOwnerReferenceKind(podList []v1.Pod, kind string) (output []v1.Pod) {
	for _, pod := range podList {
		for _, ownerReference := range pod.ObjectMeta.OwnerReferences {
			if ownerReference.Kind != kind {
				output = append(output, pod)
			}
		}
	}

	return output
}

//getPodsToBeDeleted gets the list od pods that needs to be deleted before
//deleting the k8s node.
func (ks KillerService) getPodsToBeDeleted(name string) ([]v1.Pod, error) {
	podList, err := ks.kubeClient.GetPodsInNode(name)

	if err != nil {
		return podList, err
	}
	// Filter out DaemonSet from the list of pods
	filteredPodsByDaemonSet := filterOutPodByOwnerReferenceKind(podList, "DaemonSet")

	podsToDelete := filterMirrorPods(filteredPodsByDaemonSet)	
	return podsToDelete, err
}

func filterMirrorPods(pods []v1.Pod) (filteredPods []v1.Pod) {

	for _, pod := range pods {
		if _, found := pod.ObjectMeta.Annotations[v1.MirrorPodAnnotationKey]; !found {
			filteredPods = append(filteredPods, pod)
		}
	}

	return filteredPods

}
