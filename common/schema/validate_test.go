package schema

import (
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/common/errors"
	"github.com/block/ftl/common/slices"
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
					export verb one(builtin.Empty) builtin.Empty
						+calls two.two
				}

				module two {
					export verb two(builtin.Empty) builtin.Empty
						+calls one.one
				}
				`,
			errs: []string{"found cycle in dependencies: two -> one -> two"}},
		{name: "ThreeModulesNoCycle",
			schema: `
				module one {
					verb one(builtin.Empty) builtin.Empty
						+calls two.two
				}

				module two {
					export verb two(builtin.Empty) builtin.Empty
						+calls three.three
				}

				module three {
					export verb three(builtin.Empty) builtin.Empty
				}
				`},
		{name: "ThreeModulesCycle",
			schema: `
				module one {
					export verb one(builtin.Empty) builtin.Empty
						+calls two.two
				}

				module two {
					export verb two(builtin.Empty) builtin.Empty
						+calls three.three
				}

				module three {
					export verb three(builtin.Empty) builtin.Empty
						+calls one.one
				}
				`,
			errs: []string{"found cycle in dependencies: two -> three -> one -> two"}},
		{name: "TwoModuleCycleDiffVerbs",
			schema: `
				module one {
					verb a(builtin.Empty) builtin.Empty
						+calls two.a
					export verb b(builtin.Empty) builtin.Empty
				}

				module two {
					export verb a(builtin.Empty) builtin.Empty
						+calls one.b
				}
				`,
			errs: []string{"found cycle in dependencies: two -> one -> two"}},
		{name: "SelfReference",
			schema: `
				module one {
					verb a(builtin.Empty) builtin.Empty
						+calls one.b

					verb b(builtin.Empty) builtin.Empty
						+calls one.a
				}
			`},
		{name: "ValidIngressRequestType",
			schema: `
				module one {
					export verb a(builtin.HttpRequest<Unit, Unit, Unit>) builtin.HttpResponse<builtin.Empty, builtin.Empty>
						+ingress http GET /a
				}
			`},
		{name: "InvalidIngressRequestType",
			schema: `
				module one {
					export verb a(builtin.Empty) builtin.Empty
						+ingress http GET /a
				}
			`,
			errs: []string{
				"3:20: ingress verb a: request type builtin.Empty must be builtin.HttpRequest",
				"3:35: ingress verb a: response type builtin.Empty must be builtin.HttpResponse",
			}},
		{name: "IngressBodyTypes",
			schema: `
				module one {
					export verb bytes(builtin.HttpRequest<Bytes, Unit, Unit>) builtin.HttpResponse<Bytes, Bytes>
						+ingress http POST /bytes
					export verb string(builtin.HttpRequest<String, Unit, Unit>) builtin.HttpResponse<String, String>
						+ingress http POST /string
					export verb data(builtin.HttpRequest<builtin.Empty, Unit, Unit>) builtin.HttpResponse<builtin.Empty, builtin.Empty>
						+ingress http POST /data

					// Invalid types.
					export verb any(builtin.HttpRequest<Any, Unit, Unit>) builtin.HttpResponse<Any, Any>
						+ingress http GET /any
					export verb path(builtin.HttpRequest<Unit, String, Unit>) builtin.HttpResponse<String, String>
						+ingress http GET /path/{invalid}
					export verb pathInvalid(builtin.HttpRequest<Unit, String, Unit>) builtin.HttpResponse<String, String>
						+ingress http GET /path/{invalid}/{extra}
					export verb pathMissing(builtin.HttpRequest<Unit, one.Path, Unit>) builtin.HttpResponse<String, String>
						+ingress http GET /path/{missing}
					export verb pathFound(builtin.HttpRequest<Unit, one.Path, Unit>) builtin.HttpResponse<String, String>
						+ingress http GET /path/{parameter}

					// Data comment
					export data Path {
						parameter String
					}
				}
			`,
			errs: []string{
				"11:22: ingress verb any: GET request type builtin.HttpRequest<Any, Unit, Unit> must have a body of unit not Any",
				"11:60: ingress verb any: response type builtin.HttpResponse<Any, Any> must have a body of bytes, string, data structure, unit, float, int, bool, map, or array not Any",
				"16:31: ingress verb pathInvalid: cannot use path parameter \"invalid\" with request type String as it has multiple path parameters, expected Data or Map type",
				"16:41: ingress verb pathInvalid: cannot use path parameter \"extra\" with request type String as it has multiple path parameters, expected Data or Map type",
				"18:31: ingress verb pathMissing: request pathParameter type one.Path does not contain a field corresponding to the parameter \"missing\"",
				"18:7: duplicate http ingress GET /path/{} for 19:6:\"pathFound\" and 17:6:\"pathMissing\"",
				"20:7: duplicate http ingress GET /path/{} for 13:6:\"path\" and 19:6:\"pathFound\"",
			}},
		{name: "GetRequestWithBody",
			schema: `
				module one {
					export verb bytes(builtin.HttpRequest<Bytes, Unit, Unit>) builtin.HttpResponse<Bytes, Bytes>
						+ingress http GET /bytes
				}
			`,
			errs: []string{
				"3:24: ingress verb bytes: GET request type builtin.HttpRequest<Bytes, Unit, Unit> must have a body of unit not Bytes",
			}},
		{name: "Array",
			schema: `
				module one {
					data Data {}
					export verb one(builtin.HttpRequest<[one.Data], Unit, Unit>) builtin.HttpResponse<[one.Data], builtin.Empty>
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
				"5:7: verb can not have multiple instances of cronjob",
			},
		},
		{name: "DoubleIngress",
			schema: `
				module one {
					data Data {}
					export verb one(builtin.HttpRequest<[one.Data], Unit, Unit>) builtin.HttpResponse<[one.Data], builtin.Empty>
					    +ingress http POST /one
					    +ingress http POST /two
				}
			`,
			errs: []string{
				"6:10: verb can not have multiple instances of ingress",
			},
		},
		{name: "CronOnNonEmptyVerb",
			schema: `
				module one {
					verb verbWithWrongInput(builtin.Empty) Unit
						+cron * * * * * * *
					verb verbWithWrongOutput(Unit) builtin.Empty
						+cron * * * * * * *
				}
			`,
			errs: []string{
				"4:7: verb verbWithWrongInput: cron job can not have a request type",
				"6:7: verb verbWithWrongOutput: cron job can not have a response type",
			},
		},
		{name: "IngressBodyExternalType",
			schema: `
				module two {
					export data Data {}
				}
				module one {
					export verb a(builtin.HttpRequest<two.Data, Unit, Unit>) builtin.HttpResponse<two.Data, builtin.Empty>
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
				`4:21: duplicate config "FTL_ENDPOINT", first defined at 3:20`,
				`5:21: duplicate config "FTL_ENDPOINT", first defined at 3:20`,
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
				`4:6: duplicate secret "MY_SECRET", first defined at 3:6`,
				`5:6: duplicate secret "MY_SECRET", first defined at 3:6`,
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
				`4:6: duplicate database "MY_DB", first defined at 3:6`,
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
			errs: []string{"4:7: enum variant \"A\" of type Int cannot have a value of type \"String\""},
		},
		{name: "NonSubscriberVerbsWithRetry",
			schema: `
				module one {
					verb A(builtin.Empty) Unit
						+retry 10 5s 20m
					verb B(builtin.Empty) Unit
						+retry 1m5s 20m30s
					verb C(builtin.Empty) Unit
				}
				`,
			errs: []string{
				`4:7: retries can only be added to subscribers`,
				`6:7: retries can only be added to subscribers`,
			},
		},
		{name: "InvalidRetryDurations",
			schema: `
				module one {

					data Event {}
				    topic topicA one.Event

					verb A(one.Event) Unit
						+retry 10 5s1m
						+subscribe one.topicA from=beginning
					verb B(one.Event) Unit
						+retry 1d1m5s1d
						+subscribe one.topicA from=beginning
					verb C(one.Event) Unit
						+retry 0h0m0s
						+subscribe one.topicA from=beginning
					verb D(one.Event) Unit
						+retry 1
						+subscribe one.topicA from=beginning
					verb E(one.Event) Unit
						+retry
						+subscribe one.topicA from=beginning
					verb F(one.Event) Unit
						+retry 20m20m
						+subscribe one.topicA from=beginning
					verb G(one.Event) Unit
						+retry 1s
						+retry 1s
						+subscribe one.topicA from=beginning
					verb H(one.Event) Unit
						+retry 2mins
						+subscribe one.topicA from=beginning
					verb I(one.Event) Unit
						+retry 1m 1s
						+subscribe one.topicA from=beginning
					verb J(one.Event) Unit
						+retry 1d1s
						+subscribe one.topicA from=beginning
					verb K(one.Event) Unit
						+retry 0 5s
						+subscribe one.topicA from=beginning

					verb catchSub(builtin.CatchRequest<Unit>) Unit

				}
				`,
			errs: []string{
				"11:7: could not parse min backoff duration: could not parse retry duration: duration has unit \"d\" out of order - units need to be ordered from largest to smallest - eg '1d3h2m'",
				"14:7: could not parse min backoff duration: retry must have a minimum backoff of 1s",
				"17:7: retry must have a minimum backoff",
				"20:7: retry must have a minimum backoff",
				"23:7: could not parse min backoff duration: could not parse retry duration: duration has unit \"m\" out of order - units need to be ordered from largest to smallest - eg '1d3h2m'",
				"27:7: verb can not have multiple instances of retry",
				"30:7: could not parse min backoff duration: could not parse retry duration: duration has unknown unit \"mins\" - use 'd', 'h', 'm' or 's', eg '1d' or '30s'",
				"33:7: max backoff duration (1s) needs to be at least as long as initial backoff (1m)",
				"36:7: could not parse min backoff duration: retry backoff can not be larger than 1d",
				"39:7: can not define a backoff duration when retry count is 0 and no catch is declared",
				"8:7: could not parse min backoff duration: could not parse retry duration: duration has unit \"m\" out of order - units need to be ordered from largest to smallest - eg '1d3h2m'",
			},
		},
		{name: "InvalidRetryInvalidSpace",
			schema: `
				module one {
					verb A(builtin.Empty) Unit
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

				export data eventA {
				}

				data eventB {
				}

				verb wrongEventType(test.eventA) Unit
					+subscribe test.topicB from=beginning

				verb SourceCantSubscribe(Unit) test.eventB
					+subscribe test.topicB from=latest

				verb EmptyCantSubscribe(Unit) Unit
					+subscribe test.topicB from=beginning
			}
			`,
			errs: []string{
				"16:6: verb wrongEventType: request type test.eventA differs from subscription's event type test.eventB",
				"19:6: verb SourceCantSubscribe: must be a sink to subscribe but found response type test.eventB",
				"19:6: verb SourceCantSubscribe: request type Unit differs from subscription's event type test.eventB",
				"22:6: verb EmptyCantSubscribe: request type Unit differs from subscription's event type test.eventB",
				`7:5: invalid name: must consist of only letters, numbers and underscores, and start with a lowercase letter.`,
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

			// catch verbs

			verb catchA(builtin.CatchRequest<test.EventA>) Unit
			verb catchB(builtin.CatchRequest<test.EventB>) Unit
			verb catchAWithResponse(builtin.CatchRequest<test.EventA>) test.EventA
			verb catchUnit(Unit) Unit
			verb catchBWithEventType(test.EventB) Unit

			// subscribers

			verb correctSubA(test.EventA) Unit
				+subscribe test.topicA from=beginning
				+retry 1 1s catch test.catchA

			verb correctSubB(test.EventB) Unit
				+subscribe test.topicB from=latest
				+retry 1 1s catch test.catchB

			verb incorrectSubAWithCatchB(test.EventA) Unit
				+subscribe test.topicA from=beginning
				+retry 1 1s catch test.catchB

			verb incorrectSubAWithCatchWithResponse(test.EventA) Unit
				+subscribe test.topicA from=beginning
				+retry 1 1s catch test.catchAWithResponse

			verb incorrectSubBWithCatchUnit(test.EventB) Unit
				+subscribe test.topicB from=beginning
				+retry 1 1s catch test.catchUnit

			verb incorrectSubBWithCatchEvent(test.EventB) Unit
				+subscribe test.topicB from=beginning
				+retry 1 1s catch test.catchBWithEventType

			verb incorrectSubBWithCatchNotAVerb(test.EventB) Unit
				+subscribe test.topicB from=beginning
				+retry 1 1s catch test.EventB
		}
		`,
			errs: []string{
				"31:5: catch verb must have a request type of builtin.CatchRequest<test.EventA> or builtin.CatchRequest<Any>, but found builtin.CatchRequest<test.EventB>",
				"35:5: catch verb must not have a response type but found test.EventA",
				"39:5: catch verb must have a request type of builtin.CatchRequest<test.EventB> or builtin.CatchRequest<Any>, but found Unit",
				"43:5: catch verb must have a request type of builtin.CatchRequest<test.EventB> or builtin.CatchRequest<Any>, but found test.EventB",
				"47:5: expected catch to be a verb",
			},
		},
		{
			name: "DoubleSubscribe",
			schema: `
			module one {
				data EventA {}

				topic topicA one.EventA

				verb subA(one.EventA) Unit
					+subscribe one.topicA from=beginning
					+subscribe one.topicA from=beginning
			}
		`,
			errs: []string{
				`9:6: verb can not subscribe to multiple topics`,
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
					export verb one(builtin.Empty) builtin.Empty
				}
				`,
			moduleSchema: `
				module two {
					export verb two(builtin.Empty) one.Test
						+calls one.one
				}`,
		},
		{name: "NonExportedVerbCall",
			schema: `
				module one {
					verb one(builtin.Empty) builtin.Empty
				}
				`,
			moduleSchema: `
				module two {
					export verb two(builtin.Empty) builtin.Empty
						+calls one.one
				}`,
			errs: []string{
				`4:14: verb "one.one" must be exported`,
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
