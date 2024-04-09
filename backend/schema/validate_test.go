package schema

import (
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/internal/errors"
	"github.com/TBD54566975/ftl/internal/slices"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name   string
		schema string
		errs   []string
	}{
		{name: "TwoModuleCycle",
			schema: `
				module one {
					verb one(Empty) Empty
						+calls two.two
				}

				module two {
					verb two(Empty) Empty
						+calls one.one
				}
				`,
			errs: []string{"found cycle in dependencies: two -> one -> two"}},
		{name: "ThreeModulesNoCycle",
			schema: `
				module one {
					verb one(Empty) Empty
						+calls two.two
				}

				module two {
					verb two(Empty) Empty
						+calls three.three
				}

				module three {
					verb three(Empty) Empty
				}
				`},
		{name: "ThreeModulesCycle",
			schema: `
				module one {
					verb one(Empty) Empty
						+calls two.two
				}

				module two {
					verb two(Empty) Empty
						+calls three.three
				}

				module three {
					verb three(Empty) Empty
						+calls one.one
				}
				`,
			errs: []string{"found cycle in dependencies: two -> three -> one -> two"}},
		{name: "TwoModuleCycleDiffVerbs",
			schema: `
				module one {
					verb a(Empty) Empty
						+calls two.a
					verb b(Empty) Empty
				}

				module two {
					verb a(Empty) Empty
						+calls one.b
				}
				`,
			errs: []string{"found cycle in dependencies: two -> one -> two"}},
		{name: "SelfReference",
			schema: `
				module one {
					verb a(Empty) Empty
						+calls one.b

					verb b(Empty) Empty
						+calls one.a
				}
			`},
		{name: "ValidIngressRequestType",
			schema: `
				module one {
					verb a(HttpRequest<Empty>) HttpResponse<Empty, Empty>
						+ingress http GET /a
				}
			`},
		{name: "InvalidIngressRequestType",
			schema: `
				module one {
					verb a(Empty) Empty
						+ingress http GET /a
				}
			`,
			errs: []string{
				"3:13-13: ingress verb a: request type Empty must be builtin.HttpRequest",
				"3:20-20: ingress verb a: response type Empty must be builtin.HttpRequest",
			}},
		{name: "IngressBodyTypes",
			schema: `
				module one {
					verb bytes(HttpRequest<Bytes>) HttpResponse<Bytes, Bytes>
						+ingress http GET /bytes
					verb string(HttpRequest<String>) HttpResponse<String, String>
						+ingress http GET /string
					verb data(HttpRequest<Empty>) HttpResponse<Empty, Empty>
						+ingress http GET /data

					// Invalid types.
					verb any(HttpRequest<Any>) HttpResponse<Any, Any>
						+ingress http GET /any
					verb path(HttpRequest<String>) HttpResponse<String, String>
						+ingress http GET /path/{invalid}
					verb pathMissing(HttpRequest<one.Path>) HttpResponse<String, String>
						+ingress http GET /path/{missing}
					verb pathFound(HttpRequest<one.Path>) HttpResponse<String, String>
						+ingress http GET /path/{parameter}

					data Path {
						parameter String
					}
				}
			`,
			errs: []string{
				"11:15-15: ingress verb any: request type HttpRequest<Any> must have a body of bytes, string, data structure, unit, float, int, bool, map, or array not Any",
				"11:33-33: ingress verb any: response type HttpResponse<Any, Any> must have a body of bytes, string, data structure, unit, float, int, bool, map, or array not Any",
				"14:31-31: ingress verb path: cannot use path parameter \"invalid\" with request type String, expected Data type",
				"16:31-31: ingress verb pathMissing: request type one.Path does not contain a field corresponding to the parameter \"missing\"",
				"16:7-7: duplicate http ingress GET /path/{} for 17:6:\"pathFound\" and 15:6:\"pathMissing\"",
				"18:7-7: duplicate http ingress GET /path/{} for 13:6:\"path\" and 17:6:\"pathFound\"",
			}},
		{name: "Array",
			schema: `
				module one {
					data Data {}
					verb one(HttpRequest<[one.Data]>) HttpResponse<[one.Data], Empty>
						+ingress http GET /one
				}
			`,
		},
		{name: "DoubleCron",
			schema: `
				module one {
					verb cronjob(Unit) Unit
						+cron * */2 0-23/2,4-5 * * * *
						+cron * * * * * * *
				}
			`,
			errs: []string{
				"5:7-7: only a single cron schedule is allowed per verb",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ParseString("", test.schema)
			if test.errs == nil {
				assert.NoError(t, err)
			} else {
				errs := slices.Map(errors.UnwrapAll(err), func(e error) string { return e.Error() })
				assert.Equal(t, test.errs, errs)
			}
		})
	}
}
