package schema

import (
	"golang.design/x/reflect"
)

// Builtins returns a [Module] containing built-in types.
func Builtins() *Module {
	return reflect.DeepCopy(&Module{
		Comments: []string{
			"Built-in types for FTL.",
		},
		Builtin: true,
		Name:    "builtin",
		Decls: []Decl{
			&Data{
				Comments: []string{
					"HTTP request structure used for HTTP ingress verbs.",
				},
				Name: "HttpRequest",
				TypeParameters: []*TypeParameter{
					{
						Name: "Body",
					},
				},
				Fields: []*Field{
					{
						Comments: []string{},
						Name:     "method",
						Type:     &String{},
					},
					{
						Comments: []string{},
						Name:     "path",
						Type:     &String{},
					},
					{
						Comments: []string{},
						Name:     "pathParameters",
						Type: &Map{
							Key:   &String{},
							Value: &String{},
						},
					},
					{
						Comments: []string{},
						Name:     "query",
						Type: &Map{
							Key: &String{},
							Value: &Array{
								Element: &String{},
							},
						},
					},
					{
						Comments: []string{},
						Name:     "headers",
						Type: &Map{
							Key: &String{},
							Value: &Array{
								Element: &String{},
							},
						},
					},
					{
						Comments: []string{},
						Name:     "body",
						Type: &TypeParameter{
							Name: "Body",
						},
					},
				},
			},
			&Data{
				Comments: []string{
					"HTTP response structure used for HTTP ingress verbs.",
				},
				Name: "HttpResponse",
				TypeParameters: []*TypeParameter{
					{
						Name: "Body",
					},
				},
				Fields: []*Field{
					{
						Comments: []string{},
						Name:     "status",
						Type:     &Int{},
					},
					{
						Comments: []string{},
						Name:     "headers",
						Type: &Map{
							Key: &String{},
							Value: &Array{
								Element: &String{},
							},
						},
					},
					{
						Comments: []string{},
						Name:     "body",
						Type: &TypeParameter{
							Name: "Body",
						},
					},
				},
			},
		},
	})
}
