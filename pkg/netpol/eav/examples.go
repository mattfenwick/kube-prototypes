package eav

// All/All
var AllSourcesAllDests = &Policy{
	TrafficMatcher: EverythingMatcher,
	Directive:      DirectiveAllow,
}

// All/Internal
var AllSourcesInternalDests = &Policy{
	TrafficMatcher: &Not{&Bool{DestinationIsExternalSelector}},
	Directive:      DirectiveAllow,
}

// All/External
var AllSourcesExternalDests = &Policy{
	TrafficMatcher: &Bool{DestinationIsExternalSelector},
	Directive:      DirectiveAllow,
}

//// All/None -- this doesn't make any sense
//var AllSourcesNoDests = &Policy{
//	TrafficMatcher: NothingMatcher,
//	Directive:      DirectiveAllow,
//}

// Internal/All
var InternalSourcesAllDests = &Policy{
	TrafficMatcher: &Not{&Bool{SourceIsExternalSelector}},
	Directive:      DirectiveAllow,
}

// Internal/Internal
var InternalSourcesInternalDests = &Policy{
	TrafficMatcher: NewAll(
		&Not{&Bool{SourceIsExternalSelector}},
		&Not{&Bool{DestinationIsExternalSelector}}),
	Directive: DirectiveAllow,
}

// Internal/External
var InternalSourcesExternalDests = &Policy{
	TrafficMatcher: NewAll(
		&Not{&Bool{SourceIsExternalSelector}},
		&Bool{DestinationIsExternalSelector}),
	Directive: DirectiveAllow,
}

// Internal/None
//var InternalSourcesNoDests = &Policy{
//	TrafficMatcher: ???,
//	Directive:      DirectiveAllow,
//}

// TODO these are probably all useless -- since they don't match any targets:
// External/All
// External/Internal
// External/External
// External/None
// None/All
// None/Internal
// None/External
// None/None

var PodLabelSourceNamespaceLabelDest = &Policy{
	TrafficMatcher: NewAll(
		&LabelMatcher{
			Selector: SourcePodLabelsSelector,
			Key:      "app",
			Value:    "web",
		},
		&LabelMatcher{
			Selector: DestinationNamespaceLabelsSelector,
			Key:      "stage",
			Value:    "dev",
		}),
	Directive: DirectiveAllow,
}

var SameNamespaceSourceAndDest = &Policy{
	TrafficMatcher: SameNamespaceMatcher,
	Directive:      DirectiveAllow,
}
