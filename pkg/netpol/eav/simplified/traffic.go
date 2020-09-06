package eav

import (
	"encoding/json"
	"github.com/mattfenwick/kube-prototypes/pkg/netpol/utils"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"reflect"
)

type Selector interface {
	//Select(t *Traffic) interface{}
	Select(tm *TrafficMap) interface{}
}

type KeyPathSelector struct {
	KeyPath []string
}

func NewKeyPathSelector(keyPath ...string) *KeyPathSelector {
	return &KeyPathSelector{KeyPath: keyPath}
}

func (kps *KeyPathSelector) Select(tm *TrafficMap) interface{} {
	return tm.MustApplyKeyPath(kps.KeyPath)
}

type ConstantSelector struct {
	Value interface{}
}

func (cs *ConstantSelector) Select(tm *TrafficMap) interface{} {
	return cs.Value
}

var (
	// Traffic selectors
	SourceSelector   = "Source"
	DestSelector     = "Destination"
	ProtocolSelector = "Protocol"
	PortSelector     = "Port"

	// Peer selectors
	InternalSelector = "Internal"
	IPSelector       = "IP"

	// InteralPeer selectors
	PodLabelsSelector       = "PodLabels"
	PodSelector             = "Pod"
	NamespaceLabelsSelector = "NamespaceLabels"
	NamespaceSelector       = "Namespace"
	NodeLabelsSelector      = "NodeLabels"
	NodeSelector            = "Node"
)

type Traffic struct {
	Source      *Peer
	Destination *Peer

	Protocol v1.Protocol
	Port     intstr.IntOrString
}

type Peer struct {
	Internal *InternalPeer
	IP       string
}

func (tc *Peer) IsExternal() bool {
	return tc.Internal == nil
}

type InternalPeer struct {
	PodLabels       map[string]string
	Pod             string
	NamespaceLabels map[string]string
	Namespace       string
	NodeLabels      map[string]string
	Node            string
}

// map analog of Traffic, for easier, data-driven traversal

type TrafficMap map[string]interface{}

// IsValid goal: detect if there's any fields not matching the Traffic schema:
//  - extra field -> return false
//  - missing field -> return false
//  - wrong type for field -> return false
func (tm TrafficMap) IsValid() bool {
	bytes, err := json.Marshal(tm)
	utils.DoOrDie(err)
	traffic := &Traffic{}
	err = json.Unmarshal(bytes, traffic)
	utils.DoOrDie(err)
	trafficBytes, err := json.Marshal(traffic)
	utils.DoOrDie(err)
	newTraffic := &Traffic{}
	err = json.Unmarshal(bytes, newTraffic)
	utils.DoOrDie(err)
	return string(trafficBytes) == string(bytes) && reflect.DeepEqual(traffic, newTraffic)
}

func (tm TrafficMap) ApplyKeyPath(keyPath []string) (interface{}, error) {
	obj := tm
	var applied []string
	for i := 0; i < len(keyPath)-1; i++ {
		key := keyPath[i]
		if obj == nil {
			return nil, errors.Errorf("obj is nil for key %s at index %d (keypath %+v)", key, i, keyPath)
		}
		if _, ok := obj[key]; !ok {
			return nil, errors.Errorf("obj does not have key %s (index %d, keypath %+v)", key, i, keyPath)
		}
		nextObj, ok := obj[key].(map[string]interface{})
		if !ok {
			return nil, errors.Errorf("expected map[string]interface{} at key %s (index %d, keypath %+v), found %T", key, i, keyPath, obj[key])
		}
		obj = nextObj
		applied = append(applied, key)
	}
	return obj, nil
}

func (tm TrafficMap) MustApplyKeyPath(keyPath []string) interface{} {
	val, err := tm.ApplyKeyPath(keyPath)
	utils.DoOrDie(err)
	return val
}
