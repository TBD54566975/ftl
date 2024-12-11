package terminal

import (
	"github.com/posener/complete"

	"github.com/TBD54566975/ftl/common/schema"
	"github.com/TBD54566975/ftl/internal/schema/schemaeventsource"
)

func Predictors(view schemaeventsource.View) map[string]complete.Predictor {
	return map[string]complete.Predictor{
		"verbs": &verbPredictor{view: view},
	}
}

type verbPredictor struct {
	view schemaeventsource.View
}

func (v *verbPredictor) Predict(args complete.Args) []string {
	sch := v.view.Get()
	ret := []string{}
	for _, module := range sch.Modules {
		for _, dec := range module.Decls {
			if verb, ok := dec.(*schema.Verb); ok {
				ref := schema.Ref{Module: module.Name, Name: verb.Name}
				ret = append(ret, ref.String())
			}
		}
	}
	return ret
}
