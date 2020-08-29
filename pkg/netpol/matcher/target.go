package matcher

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sort"
)

// Target represents a NetworkPolicySpec.PodSelector, which is in a namespace
type Target struct {
	Namespace   string
	PodSelector metav1.LabelSelector
	Egress      *TrafficPeers
	Ingress     *TrafficPeers
	primaryKey  string
}

// Combine creates a new Target combining the egress and ingress rules
// of the two original targets.  Neither input is modified.
// The Primary Keys of the two targets must match.
func (t *Target) Combine(other *Target) *Target {
	myPk := t.GetPrimaryKey()
	otherPk := other.GetPrimaryKey()
	if myPk != otherPk {
		panic(errors.Errorf("cannot combine targets: primary keys differ -- '%s' vs '%s'", myPk, otherPk))
	}

	return &Target{
		Namespace:   t.Namespace,
		PodSelector: t.PodSelector,
		Egress:      t.Egress.Combine(other.Egress),
		Ingress:     t.Ingress.Combine(other.Ingress),
	}
}

// The primary key is a combination of PodSelector and namespace
func (t *Target) GetPrimaryKey() string {
	if t.primaryKey == "" {
		var labelKeys []string
		for key := range t.PodSelector.MatchLabels {
			labelKeys = append(labelKeys, key)
		}
		sort.Slice(labelKeys, func(i, j int) bool {
			return labelKeys[i] < labelKeys[j]
		})
		var keyVals []string
		for _, key := range labelKeys {
			keyVals = append(keyVals, fmt.Sprintf("%s: %s", key, t.PodSelector.MatchLabels))
		}
		// this is weird, but use an array to make the order deterministic
		bytes, err := json.Marshal([]interface{}{"Namespace", t.Namespace, "MatchLabels", keyVals, "MatchExpression", t.PodSelector.MatchExpressions})
		if err != nil {
			log.Fatalf("%+v", err)
		}
		t.primaryKey = string(bytes)
	}
	return t.primaryKey
}

func (t *Target) Allows(td *TrafficDirection) bool {
	if td.IsIngress {
		if t.Ingress == nil {
			return false
		}
		return t.Ingress.Allows(td)
	} else {
		if t.Egress == nil {
			return false
		}
		return t.Egress.Allows(td)
	}
}