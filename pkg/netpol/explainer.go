package netpol

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	networkingv1 "k8s.io/api/networking/v1"
	"strings"
)

type ExplanationTarget struct {
	Namespace string
	Pods      string
}

type ExplanationRule struct {
	Pods       string
	Namespaces string
}

type ExplanationIngress struct {
	Rules []*ExplanationRule
}

func (ei *ExplanationIngress) PrettyPrint() string {
	if len(ei.Rules) == 0 {
		return "all pods and all namespaces are allowed"
	}
	var allowed []string
	for _, rule := range ei.Rules {
		allowed = append(allowed, fmt.Sprintf("(%s AND %s)", rule.Namespaces, rule.Pods))
	}
	return strings.Join(allowed, " OR ")
}

type Explanation struct {
	Target  *ExplanationTarget
	Ingress []*ExplanationIngress
}

func (e *Explanation) PrettyPrint() string {
	out := []string{
		"target namespace: " + e.Target.Namespace,
		"target pods: " + e.Target.Pods,
		"ingress:",
	}
	if len(e.Ingress) == 0 {
		out = append(out, " - no ingress is allowed")
	} else {
		var ingresses []string
		for _, ingress := range e.Ingress {
			ingresses = append(ingresses, " - "+ingress.PrettyPrint())
		}
		out = append(out, strings.Join(ingresses, " OR "))
	}
	return strings.Join(out, "\n")
}

func ExplainPolicy(policy *networkingv1.NetworkPolicy) *Explanation {
	log.Warnf("TODO policy.spec.podselector.MatchExpressions is currently ignored")
	log.Warnf("TODO look at PolicyTypes -- nil/empty = Ingress, otherwise as given")

	targetLabels := policy.Spec.PodSelector.MatchLabels
	var targetPods string
	if len(targetLabels) == 0 {
		targetPods = "all pods"
	} else {
		targetPods = fmt.Sprintf("pods with labels %s", labelString(targetLabels))
	}

	return &Explanation{
		Target: &ExplanationTarget{
			Namespace: policy.Namespace,
			Pods:      targetPods,
		},
		Ingress: ExplainIngress(policy),
	}
}

func ExplainIngress(policy *networkingv1.NetworkPolicy) []*ExplanationIngress {
	var ingresses []*ExplanationIngress
	for _, rule := range policy.Spec.Ingress {
		log.Warnf("TODO ingress ports are currently ignored")
		var rules []*ExplanationRule
		for _, from := range rule.From {
			log.Warnf("TODO ingress/from/ipblock is currently ignored")
			var pods, ns string

			if from.PodSelector == nil {
				log.Warnf("TODO this is probably worth a test -- when pod selector isn't specified, does that mean 'all pods'?")
				pods = "all pods"
			} else if len(from.PodSelector.MatchLabels) == 0 {
				pods = "all pods"
			} else {
				pods = fmt.Sprintf("only pods with labels %s", labelString(from.PodSelector.MatchLabels))
			}

			if from.NamespaceSelector == nil {
				ns = fmt.Sprintf("only namespace %s", policy.Namespace)
			} else if len(from.NamespaceSelector.MatchLabels) == 0 {
				ns = "all namespaces"
			} else {
				ns = fmt.Sprintf("only namespaces with labels %s", labelString(from.NamespaceSelector.MatchLabels))
			}

			rules = append(rules, &ExplanationRule{
				Pods:       pods,
				Namespaces: ns,
			})
		}
		ingresses = append(ingresses, &ExplanationIngress{Rules: rules})
	}
	return ingresses
}
