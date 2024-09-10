package infra

import "log"

type Matcher func(node *ResourceNode, graph *ResourceGraph) bool

type RewriteRule struct {
	Name    string
	Matcher Matcher
	Handler func(hits []*ResourceNode, graph *ResourceGraph) (string, error)
}

type RuleSet []RewriteRule

func (r *RuleSet) ApplyTo(graph *ResourceGraph) string {
	// TODO: We should only return a window of direct upstream, and downstream nodes for modification
	for _, rule := range *r {
		match := graph.Find(rule.Matcher)
		if len(match) > 0 {
			res, err := rule.Handler(match, graph)
			if err != nil {
				log.Fatal(err)
			}
			return rule.Name + " " + res
		}
	}
	return ""
}

func (r *RuleSet) ApplyAll(graph *ResourceGraph) []string {
	var result []string
	last := r.ApplyTo(graph)
	for last != "" {
		result = append(result, last)
		last = r.ApplyTo(graph)
	}
	return result
}

// Matchers
var NextToDeploy = func(node *ResourceNode, graph *ResourceGraph) bool {
	for _, out := range graph.Out(node) {
		if out.To.Properties["state"] != "ready" {
			return false
		}
	}

	return node.Properties["state"] == "planning"
}

func ByModule(name string) Matcher {
	return func(node *ResourceNode, graph *ResourceGraph) bool {
		return node.Properties["module"] == name
	}
}

// Rules
var Rules = RuleSet{{
	Name:    "provision",
	Matcher: NextToDeploy,
	Handler: func(hits []*ResourceNode, graph *ResourceGraph) (string, error) {
		// TODO: start provisioner, and transition to "provisioning"
		hits[0].Properties["state"] = "ready"
		return hits[0].Id(), nil
	},
}}
