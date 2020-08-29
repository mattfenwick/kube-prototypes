package matcher

type ResolvedTraffic struct {
	Traffic *Traffic

	// TODO what about modelling a selector-match -- if traffic would get through to them/it?
	//    - and ignore the details of how that would get resolved to a pod?
	//PodSelector metav1.LabelSelector

	Target *ResolvedPodTarget
}

type ResolvedPodTarget struct {
	PodLabels       map[string]string
	NamespaceLabels map[string]string
	Namespace       string
}
