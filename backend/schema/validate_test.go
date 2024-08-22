package schema

import (
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/internal/errors"
	"github.com/TBD54566975/ftl/internal/slices"
)

//nolint:maintidx
func TestValidate(t *testing.T) {
	tests := []struct {
		name   string
		schema string
		errs   []string
	}{
		{name: "TwoModuleCycle",
			schema: `
				module one {
					export verb one(Empty) Empty
						+calls two.two
				}

				module two {
					export verb two(Empty) Empty
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
					export verb two(Empty) Empty
						+calls three.three
				}

				module three {
					export verb three(Empty) Empty
				}
				`},
		{name: "ThreeModulesCycle",
			schema: `
				module one {
					export verb one(Empty) Empty
						+calls two.two
				}

				module two {
					export verb two(Empty) Empty
						+calls three.three
				}

				module three {
					export verb three(Empty) Empty
						+calls one.one
				}
				`,
			errs: []string{"found cycle in dependencies: two -> three -> one -> two"}},
		{name: "TwoModuleCycleDiffVerbs",
			schema: `
				module one {
					verb a(Empty) Empty
						+calls two.a
					export verb b(Empty) Empty
				}

				module two {
					export verb a(Empty) Empty
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
					export verb a(HttpRequest<Unit, Unit, Unit>) HttpResponse<Empty, Empty>
						+ingress http GET /a
				}
			`},
		{name: "InvalidIngressRequestType",
			schema: `
				module one {
					export verb a(Empty) Empty
						+ingress http GET /a
				}
			`,
			errs: []string{
				"3:20-20: ingress verb a: request type Empty must be builtin.HttpRequest",
				"3:27-27: ingress verb a: response type Empty must be builtin.HttpResponse",
			}},
		{name: "IngressBodyTypes",
			schema: `
				module one {
					export verb bytes(HttpRequest<Bytes, Unit, Unit>) HttpResponse<Bytes, Bytes>
						+ingress http POST /bytes
					export verb string(HttpRequest<String, Unit, Unit>) HttpResponse<String, String>
						+ingress http POST /string
					export verb data(HttpRequest<Empty, Unit, Unit>) HttpResponse<Empty, Empty>
						+ingress http POST /data

					// Invalid types.
					export verb any(HttpRequest<Any, Unit, Unit>) HttpResponse<Any, Any>
						+ingress http GET /any
					export verb path(HttpRequest<Unit, String, Unit>) HttpResponse<String, String>
						+ingress http GET /path/{invalid}
					export verb pathInvalid(HttpRequest<Unit, String, Unit>) HttpResponse<String, String>
						+ingress http GET /path/{invalid}/{extra}
					export verb pathMissing(HttpRequest<Unit, one.Path, Unit>) HttpResponse<String, String>
						+ingress http GET /path/{missing}
					export verb pathFound(HttpRequest<Unit, one.Path, Unit>) HttpResponse<String, String>
						+ingress http GET /path/{parameter}

					// Data comment
					export data Path {
						parameter String
					}
				}
			`,
			errs: []string{
				"11:22-22: ingress verb any: GET request type HttpRequest<Any, Unit, Unit> must have a body of unit not Any",
				"11:52-52: ingress verb any: response type HttpResponse<Any, Any> must have a body of bytes, string, data structure, unit, float, int, bool, map, or array not Any",
				"16:31-31: ingress verb pathInvalid: cannot use path parameter \"invalid\" with request type String as it has multiple path parameters, expected Data or Map type",
				"16:41-41: ingress verb pathInvalid: cannot use path parameter \"extra\" with request type String as it has multiple path parameters, expected Data or Map type",
				"18:31-31: ingress verb pathMissing: request pathParameter type one.Path does not contain a field corresponding to the parameter \"missing\"",
				"18:7-7: duplicate http ingress GET /path/{} for 19:6:\"pathFound\" and 17:6:\"pathMissing\"",
				"20:7-7: duplicate http ingress GET /path/{} for 13:6:\"path\" and 19:6:\"pathFound\"",
			}},
		{name: "GetRequestWithBody",
			schema: `
				module one {
					export verb bytes(HttpRequest<Bytes, Unit, Unit>) HttpResponse<Bytes, Bytes>
						+ingress http GET /bytes
				}
			`,
			errs: []string{
				"3:24-24: ingress verb bytes: GET request type HttpRequest<Bytes, Unit, Unit> must have a body of unit not Bytes",
			}},
		{name: "Array",
			schema: `
				module one {
					data Data {}
					export verb one(HttpRequest<[one.Data], Unit, Unit>) HttpResponse<[one.Data], Empty>
						+ingress http POST /one
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
				"5:7-7: verb can not have multiple instances of cronjob",
			},
		},
		{name: "DoubleIngress",
			schema: `
				module one {
					data Data {}
					export verb one(HttpRequest<[one.Data], Unit, Unit>) HttpResponse<[one.Data], Empty>
					    +ingress http POST /one
					    +ingress http POST /two
				}
			`,
			errs: []string{
				"6:10-10: verb can not have multiple instances of ingress",
			},
		},
		{name: "CronOnNonEmptyVerb",
			schema: `
				module one {
					verb verbWithWrongInput(Empty) Unit
						+cron * * * * * * *
					verb verbWithWrongOutput(Unit) Empty
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
					export data Data {}
				}
				module one {
					export verb a(HttpRequest<two.Data, Unit, Unit>) HttpResponse<two.Data, Empty>
						+ingress http GET /a
				}
			`,
		},
		{name: "DuplicateConfigs",
			schema: `
				module one {
                  	config FTL_ENDPOINT String
                    config FTL_ENDPOINT Any
                    config FTL_ENDPOINT String
				}
			`,
			errs: []string{
				`4:21-21: duplicate config "FTL_ENDPOINT", first defined at 3:20`,
				`5:21-21: duplicate config "FTL_ENDPOINT", first defined at 3:20`,
			},
		},
		{name: "DuplicateSecrets",
			schema: `
				module one {
					secret MY_SECRET String
					secret MY_SECRET Any
					secret MY_SECRET String
				}
			`,
			errs: []string{
				`4:6-6: duplicate secret "MY_SECRET", first defined at 3:6`,
				`5:6-6: duplicate secret "MY_SECRET", first defined at 3:6`,
			},
		},
		{name: "ConfigAndSecretsWithSameName",
			schema: `
				module one {
					config FTL_ENDPOINT String
					secret FTL_ENDPOINT String
				}
			`,
		},
		{name: "DuplicateDatabases",
			schema: `
				module one {
					database postgres MY_DB
					database postgres MY_DB
				}
			`,
			errs: []string{
				`4:6-6: duplicate database "MY_DB", first defined at 3:6`,
			},
		},
		{name: "ValueEnumMismatchedVariantTypes",
			schema: `
				module one {
					enum Enum: Int {
						A = "A"
						B = 1
					}
				}
				`,
			errs: []string{"4:7-7: enum variant \"A\" of type Int cannot have a value of type \"String\""},
		},
		{name: "InvalidFSM",
			schema: `
				module one {
					verb A(Empty) Unit
					verb B(one.C) Empty

					fsm FSM {
						transition one.C to one.B
					}
				}
				`,
			errs: []string{
				`4:13-13: unknown reference "one.C", is the type annotated and exported?`,
				`6:6-6: "FSM" has no start states`,
				`7:18-18: unknown source verb "one.C"`,
				`7:27-27: destination state "one.B" must be a sink but is verb`,
			},
		},
		{name: "DuplicateFSM",
			schema: `
				module one {
					verb A(Empty) Unit
						+retry 10 5s 20m
					verb B(Empty) Unit
						+retry 1m5s 20m30s
					verb C(Empty) Unit

					fsm FSM {
						start one.A
						transition one.A to one.B
					}

					fsm FSM {
						start one.A
						transition one.A to one.B
					}
				}
				`,
			errs: []string{
				`14:6-6: duplicate fsm "FSM", first defined at 9:6`,
			},
		},
		{name: "NonFSMVerbsWithRetry",
			schema: `
				module one {
					verb A(Empty) Unit
						+retry 10 5s 20m
					verb B(Empty) Unit
						+retry 1m5s 20m30s
					verb C(Empty) Unit
				}
				`,
			errs: []string{
				`4:7-7: retries can only be added to subscribers or FSM transitions`,
				`6:7-7: retries can only be added to subscribers or FSM transitions`,
			},
		},
		{name: "InvalidRetryDurations",
			schema: `
				module one {
					verb A(Empty) Unit
						+retry 10 5s1m
					verb B(Empty) Unit
						+retry 1d1m5s1d
					verb C(Empty) Unit
						+retry 0h0m0s
					verb D(Empty) Unit
						+retry 1
					verb E(Empty) Unit
						+retry
					verb F(Empty) Unit
						+retry 20m20m
					verb G(Empty) Unit
						+retry 1s
						+retry 1s
					verb H(Empty) Unit
						+retry 2mins
					verb I(Empty) Unit
						+retry 1m 1s
					verb J(Empty) Unit
						+retry 1d1s
					verb K(Empty) Unit
						+retry 0 5s

					verb catchFSM(builtin.CatchRequest<Unit>) Unit

					fsm FSM
						+retry 0 5s catch catchFSM
					{
						start one.A
						transition one.A to one.B
						transition one.A to one.C
						transition one.A to one.D
						transition one.A to one.E
						transition one.A to one.F
						transition one.A to one.G
						transition one.A to one.H
						transition one.A to one.I
						transition one.A to one.J
						transition one.A to one.K
					}
				}
				`,
			errs: []string{
				`10:7-7: retry must have a minimum backoff`,
				`12:7-7: retry must have a minimum backoff`,
				`14:7-7: could not parse min backoff duration: could not parse retry duration: duration has unit "m" out of order - units need to be ordered from largest to smallest - eg '1d3h2m'`,
				`17:7-7: verb can not have multiple instances of retry`,
				`19:7-7: could not parse min backoff duration: could not parse retry duration: duration has unknown unit "mins" - use 'd', 'h', 'm' or 's', eg '1d' or '30s'`,
				`21:7-7: max backoff duration (1s) needs to be at least as long as initial backoff (1m)`,
				`23:7-7: could not parse min backoff duration: retry backoff can not be larger than 1d`,
				`25:7-7: can not define a backoff duration when retry count is 0 and no catch is declared`,
				`30:7-7: catch can only be defined on verbs`,
				`4:7-7: could not parse min backoff duration: could not parse retry duration: duration has unit "m" out of order - units need to be ordered from largest to smallest - eg '1d3h2m'`,
				`6:7-7: could not parse min backoff duration: could not parse retry duration: duration has unit "d" out of order - units need to be ordered from largest to smallest - eg '1d3h2m'`,
				`8:7-7: could not parse min backoff duration: retry must have a minimum backoff of 1s`,
			},
		},
		{name: "InvalidRetryInvalidSpace",
			schema: `
				module one {
					verb A(Empty) Unit
						+retry 10 5 s
				}
				`,
			errs: []string{
				`4:19: unexpected token "s"`,
			},
		},
		{name: "InvalidPubSub",
			schema: `
			module test {
				export topic topicA test.eventA
			
				topic topicB test.eventB

				topic StartsWithUpperCase test.eventA

				subscription subA test.topicA

				subscription subB test.topicB

				export data eventA {
				}

				data eventB {
				}

				verb wrongEventType(test.eventA) Unit
					+subscribe subB

				verb SourceCantSubscribe(Unit) test.eventB
					+subscribe subB

				verb EmptyCantSubscribe(Unit) Unit
					+subscribe subB
			}
			`,
			errs: []string{
				`20:6-6: verb wrongEventType: request type test.eventA differs from subscription's event type test.eventB`,
				`23:6-6: verb SourceCantSubscribe: must be a sink to subscribe but found response type test.eventB`,
				`23:6-6: verb SourceCantSubscribe: request type Unit differs from subscription's event type test.eventB`,
				`26:6-6: verb EmptyCantSubscribe: request type Unit differs from subscription's event type test.eventB`,
				`7:5-5: invalid name: must consist of only letters, numbers and underscores, and start with a lowercase letter.`,
			},
		},
		{
			name: "PubSubCatch",
			schema: `
		module test {
			// pub sub basic set up

			data EventA {}
			data EventB {}

			topic topicA test.EventA
			topic topicB test.EventB

			subscription subA test.topicA
			subscription subB test.topicB

			// catch verbs

			verb catchA(builtin.CatchRequest<test.EventA>) Unit	
			verb catchB(builtin.CatchRequest<test.EventB>) Unit
			verb catchAWithResponse(builtin.CatchRequest<test.EventA>) test.EventA
			verb catchUnit(Unit) Unit
			verb catchBWithEventType(test.EventB) Unit

			// subscribers

			verb correctSubA(test.EventA) Unit
				+subscribe subA
				+retry 1 1s catch test.catchA

			verb correctSubB(test.EventB) Unit
				+subscribe subB
				+retry 1 1s catch test.catchB

			verb incorrectSubAWithCatchB(test.EventA) Unit
				+subscribe subA
				+retry 1 1s catch test.catchB

			verb incorrectSubAWithCatchWithResponse(test.EventA) Unit
				+subscribe subA
				+retry 1 1s catch test.catchAWithResponse

			verb incorrectSubBWithCatchUnit(test.EventB) Unit
				+subscribe subB
				+retry 1 1s catch test.catchUnit

			verb incorrectSubBWithCatchEvent(test.EventB) Unit
				+subscribe subB
				+retry 1 1s catch test.catchBWithEventType

			verb incorrectSubBWithCatchNotAVerb(test.EventB) Unit
				+subscribe subB
				+retry 1 1s catch test.EventB
		}
		`,
			errs: []string{
				"34:5-5: catch verb must have a request type of builtin.CatchRequest<test.EventA> or builtin.CatchRequest<Any>, but found builtin.CatchRequest<test.EventB>",
				"38:5-5: catch verb must not have a response type but found test.EventA",
				"42:5-5: catch verb must have a request type of builtin.CatchRequest<test.EventB> or builtin.CatchRequest<Any>, but found Unit",
				"46:5-5: catch verb must have a request type of builtin.CatchRequest<test.EventB> or builtin.CatchRequest<Any>, but found test.EventB",
				"50:5-5: expected catch to be a verb",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ParseString("", test.schema)
			if test.errs == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err, "expected errors:\n%v", strings.Join(test.errs, "\n"))
				errs := slices.Map(errors.UnwrapAll(err), func(e error) string { return e.Error() })
				assert.Equal(t, test.errs, errs)
			}
		})
	}
}

func TestValidateModuleWithSchema(t *testing.T) {
	tests := []struct {
		name         string
		schema       string
		moduleSchema string
		errs         []string
	}{
		{name: "ValidModuleWithSchema",
			schema: `
				module one {
					export data Test {}
					export verb one(Empty) Empty
				}
				`,
			moduleSchema: `
				module two {
					export verb two(Empty) one.Test
						+calls one.one
				}`,
		},
		{name: "NonExportedVerbCall",
			schema: `
				module one {
					verb one(Empty) Empty
				}
				`,
			moduleSchema: `
				module two {
					export verb two(Empty) Empty
						+calls one.one
				}`,
			errs: []string{
				`4:14-14: verb "one.one" must be exported`,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sch, err := ParseString("", test.schema)
			assert.NoError(t, err)
			module, err := ParseModuleString("", test.moduleSchema)
			assert.NoError(t, err)
			sch.Modules = append(sch.Modules, module)
			_, err = ValidateModuleInSchema(sch, optional.Some[*Module](module))
			if test.errs == nil {
				assert.NoError(t, err)
			} else {
				errs := slices.Map(errors.UnwrapAll(err), func(e error) string { return e.Error() })
				assert.Equal(t, test.errs, errs)
			}
		})
	}
}
