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
					internal verb one(Empty) Empty
						+calls two.two
				}

				module two {
					// some comments
					internal verb two(Empty) Empty
						+calls one.one
				}
				`,
			errs: []string{"found cycle in dependencies: two -> one -> two"}},
		{name: "ThreeModulesNoCycle",
			schema: `
				module one {
					internal verb one(Empty) Empty
						+calls two.two
				}

				module two {
					internal verb two(Empty) Empty
						+calls three.three
				}

				module three {
					internal verb three(Empty) Empty
				}
				`},
		{name: "ThreeModulesCycle",
			schema: `
				module one {
					internal verb one(Empty) Empty
						+calls two.two
				}

				module two {
					internal verb two(Empty) Empty
						+calls three.three
				}

				module three {
					internal verb three(Empty) Empty
						+calls one.one
				}
				`,
			errs: []string{"found cycle in dependencies: two -> three -> one -> two"}},
		{name: "TwoModuleCycleDiffVerbs",
			schema: `
				module one {
					internal verb a(Empty) Empty
						+calls two.a
					internal verb b(Empty) Empty
				}

				module two {
					internal verb a(Empty) Empty
						+calls one.b
				}
				`,
			errs: []string{"found cycle in dependencies: two -> one -> two"}},
		{name: "SelfReference",
			schema: `
				module one {
					internal verb a(Empty) Empty
						+calls one.b

					internal verb b(Empty) Empty
						+calls one.a
				}
			`},
		{name: "ValidIngressRequestType",
			schema: `
				module one {
					internal verb a(HttpRequest<Empty>) HttpResponse<Empty, Empty>
						+ingress http GET /a
				}
			`},
		{name: "InvalidIngressRequestType",
			schema: `
				module one {
					internal verb a(Empty) Empty
						+ingress http GET /a
				}
			`,
			errs: []string{
				"3:22-22: ingress verb a: request type Empty must be builtin.HttpRequest",
				"3:29-29: ingress verb a: response type Empty must be builtin.HttpRequest",
			}},
		{name: "IngressBodyTypes",
			schema: `
				module one {
					internal verb bytes(HttpRequest<Bytes>) HttpResponse<Bytes, Bytes>
						+ingress http GET /bytes
					internal verb string(HttpRequest<String>) HttpResponse<String, String>
						+ingress http GET /string
					internal verb data(HttpRequest<Empty>) HttpResponse<Empty, Empty>
						+ingress http GET /data

					// Invalid types.
					internal verb any(HttpRequest<Any>) HttpResponse<Any, Any>
						+ingress http GET /any
					internal verb path(HttpRequest<String>) HttpResponse<String, String>
						+ingress http GET /path/{invalid}
					internal verb pathMissing(HttpRequest<one.Path>) HttpResponse<String, String>
						+ingress http GET /path/{missing}
					internal verb pathFound(HttpRequest<one.Path>) HttpResponse<String, String>
						+ingress http GET /path/{parameter}

					data Path {
						parameter String
					}
				}
			`,
			errs: []string{
				"11:24-24: ingress verb any: request type HttpRequest<Any> must have a body of bytes, string, data structure, unit, float, int, bool, map, or array not Any",
				"11:42-42: ingress verb any: response type HttpResponse<Any, Any> must have a body of bytes, string, data structure, unit, float, int, bool, map, or array not Any",
				"14:31-31: ingress verb path: cannot use path parameter \"invalid\" with request type String, expected Data type",
				"16:31-31: ingress verb pathMissing: request type one.Path does not contain a field corresponding to the parameter \"missing\"",
				"16:7-7: duplicate http ingress GET /path/{} for 17:6:\"pathFound\" and 15:6:\"pathMissing\"",
				"18:7-7: duplicate http ingress GET /path/{} for 13:6:\"path\" and 17:6:\"pathFound\"",
			}},
		{name: "Array",
			schema: `
				module one {
					data Data {}
					internal verb one(HttpRequest<[one.Data]>) HttpResponse<[one.Data], Empty>
						+ingress http GET /one
				}
			`,
		},
		{name: "DoubleCron",
			schema: `
				module one {
					internal verb cronjob(Unit) Unit
						+cron * */2 0-23/2,4-5 * * * *
						+cron * * * * * * *
				}
			`,
			errs: []string{
				"5:7-7: verb can not have multiple instances of cronjob",
			},
		},
		{name: "DoubleIngress",
			schema: `
				module one {
					data Data {}
					internal verb one(HttpRequest<[one.Data]>) HttpResponse<[one.Data], Empty>
					    +ingress http GET /one
					    +ingress http GET /two
				}
			`,
			errs: []string{
				"6:10-10: verb can not have multiple instances of ingress",
			},
		},
		{name: "CronOnNonEmptyVerb",
			schema: `
				module one {
					internal verb verbWithWrongInput(Empty) Unit
						+cron * * * * * * *
					internal verb verbWithWrongOutput(Unit) Empty
						+cron * * * * * * *
				}
			`,
			errs: []string{
				"4:7-7: verb verbWithWrongInput: cron job can not have a request type",
				"6:7-7: verb verbWithWrongOutput: cron job can not have a response type",
			},
		},
		{name: "IngressBodyExternalType",
			schema: `
				module two {
					data Data {}
				}
				module one {
					internal verb a(HttpRequest<two.Data>) HttpResponse<two.Data, Empty>
						+ingress http GET /a
				}
			`},
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
