package matcher

type TrafficDirection struct {
	IsIngress  bool
	SourceDest TrafficSourceDest
	Port       *PortProtocol
	IP         string

	// TODO what about modelling a selector-match -- if traffic would get through to them/it?
	//    - and ignore the details of how that would get resolved to a pod?
	//PodSelector metav1.LabelSelector

	Target *PodTarget
}

// TODO does this make sense?
//   this models a request that has been resolved to a pod -- will it get through?
type PodTarget struct {
	// the labels on the target pod
	PodLabels map[string]string
	// TODO this doesn't make sense, right?
	//// the labels on the namespace of the target pod
	//NamespaceLabels map[string]string
	// TODO do we need this?
	Namespace string
	// TODO do we need this?
	//TargetPod string
}

type TrafficSourceDest interface {
	IsExternal() bool
	GetPodLabels() map[string]string
	GetNamespaceLabels() map[string]string
	GetNamespace() string
}

// PodTraffic represents traffic whose source/dest is a pod in the same cluster.
type PodTraffic struct {
	PodLabels       map[string]string
	NamespaceLabels map[string]string
	Namespace       string
}

func (pt *PodTraffic) IsExternal() bool {
	return false
}

func (pt *PodTraffic) GetPodLabels() map[string]string {
	return pt.PodLabels
}

func (pt *PodTraffic) GetNamespaceLabels() map[string]string {
	return pt.NamespaceLabels
}

func (pt *PodTraffic) GetNamespace() string {
	return pt.Namespace
}

// ExternalTraffic represents traffic whose source/dest is external to the cluster
type ExternalTraffic struct{}

func (et *ExternalTraffic) IsExternal() bool {
	return true
}

func (et *ExternalTraffic) GetPodLabels() map[string]string {
	// TODO should this be an empty map?
	return nil
}

func (et *ExternalTraffic) GetNamespaceLabels() map[string]string {
	// TODO should this be an empty map?
	return nil
}

func (et *ExternalTraffic) GetNamespace() string {
	return ""
}
