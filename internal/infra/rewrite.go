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
func NextToDeploy(kind string) Matcher {
	return func(node *ResourceNode, graph *ResourceGraph) bool {
		for _, out := range graph.Out(node) {
			if out.To.Properties["state"] != "ready" {
				return false
			}
		}

		return node.Properties["state"] == "planning" && node.Properties["kind"] == kind
	}
}

func NextToRemove(kind string) Matcher {
	return func(node *ResourceNode, graph *ResourceGraph) bool {
		for _, in := range graph.In(node) {
			if in.From.Properties["kind"] != "FTL" && in.From.Properties["state"] != "removed" {
				return false
			}
		}

		return node.Properties["state"] == "outdated" && node.Properties["kind"] == kind
	}
}

func NextToMarkOutdated() Matcher {
	return func(node *ResourceNode, graph *ResourceGraph) bool {
		incoming := graph.In(node)
		isOutdated := len(incoming) > 0
		for _, in := range incoming {
			if in.From.Properties["state"] != "removed" {
				isOutdated = false
			}
		}

		return node.Properties["state"] == "ready" && isOutdated
	}
}

func NextToClear() Matcher {
	return func(node *ResourceNode, graph *ResourceGraph) bool {
		return node.Properties["state"] == "removed" && len(graph.Out(node)) == 0
	}
}

func ByModule(name string) Matcher {
	return func(node *ResourceNode, graph *ResourceGraph) bool {
		return node.Properties["module"] == name
	}
}

// Rules
var Rules = RuleSet{{
	Name:    "run provision-module",
	Matcher: NextToDeploy("module"),
	Handler: func(hits []*ResourceNode, graph *ResourceGraph) (string, error) {
		hits[0].Properties["state"] = "ready"
		return hits[0].Id(), nil
	},
}, {
	Name:    "run provision-database",
	Matcher: NextToDeploy("database"),
	Handler: func(hits []*ResourceNode, graph *ResourceGraph) (string, error) {
		hits[0].Properties["state"] = "ready"
		return hits[0].Id(), nil
	},
}, {
	Name:    "run delete-database",
	Matcher: NextToRemove("database"),
	Handler: func(hits []*ResourceNode, graph *ResourceGraph) (string, error) {
		hits[0].Properties["state"] = "removed"
		return hits[0].Id(), nil
	},
}, {
	Name:    "run delete-module",
	Matcher: NextToRemove("module"),
	Handler: func(hits []*ResourceNode, graph *ResourceGraph) (string, error) {
		hits[0].Properties["state"] = "removed"
		return hits[0].Id(), nil
	},
}, {
	Name:    "outdate",
	Matcher: NextToMarkOutdated(),
	Handler: func(hits []*ResourceNode, graph *ResourceGraph) (string, error) {
		hits[0].Properties["state"] = "outdated"
		return hits[0].Id(), nil
	},
}, {
	Name:    "delete",
	Matcher: NextToClear(),
	Handler: func(hits []*ResourceNode, graph *ResourceGraph) (string, error) {
		id := hits[0].Id()
		graph.DelNode(hits[0])
		return id, nil
	},
}}
