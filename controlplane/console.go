package controlplane

import (
	"context"

	"github.com/alecthomas/errors"
	"github.com/bufbuild/connect-go"

	"github.com/TBD54566975/ftl/controlplane/internal/dal"
	"github.com/TBD54566975/ftl/internal/slices"
	"github.com/TBD54566975/ftl/observability"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	pbconsole "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/console"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/console/pbconsoleconnect"
	pschema "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/schema"
)

type ConsoleService struct {
	dal *dal.DAL
}

var _ pbconsoleconnect.ConsoleServiceHandler = (*ConsoleService)(nil)

func NewConsoleService(dal *dal.DAL) *ConsoleService {
	return &ConsoleService{
		dal: dal,
	}
}

func (*ConsoleService) Ping(context.Context, *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

func (c *ConsoleService) GetModules(ctx context.Context, req *connect.Request[pbconsole.GetModulesRequest]) (*connect.Response[pbconsole.GetModulesResponse], error) {
	deployments, err := c.dal.GetActiveDeployments(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	moduleNames, err := slices.MapErr(deployments, func(in dal.Deployment) (string, error) {
		return in.Module, nil
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	metrics, err := c.dal.GetLatestModuleMetrics(ctx, moduleNames)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var modules []*pbconsole.Module
	for _, deployment := range deployments {
		verbs := []*pbconsole.Verb{}
		data := []*pschema.Data{}
		for _, decl := range deployment.Schema.Decls {

			switch decl := decl.(type) {
			case *schema.Verb:
				//nolint:forcetypeassert
				verbs = append(verbs, &pbconsole.Verb{
					Verb:           decl.ToProto().(*pschema.Verb),
					CallCount:      getCounterMetric(metrics, deployment.Module, decl.Name, observability.CallRequestCount),
					CallLatency:    getHistogramMetric(metrics, deployment.Module, decl.Name, observability.CallLatency),
					CallErrorCount: getCounterMetric(metrics, deployment.Module, decl.Name, observability.CallErrorCount),
				})
			case *schema.Data:
				//nolint:forcetypeassert
				data = append(data, decl.ToProto().(*pschema.Data))
			}
		}
		modules = append(modules, &pbconsole.Module{
			Name:     deployment.Module,
			Language: deployment.Language,
			Verbs:    verbs,
			Data:     data,
		})
	}

	return connect.NewResponse(&pbconsole.GetModulesResponse{
		Modules: modules,
	}), nil
}

func getCounterMetric(metrics []dal.Metric, module string, verb string, metricName string) *pbconsole.MetricCounter {
	for _, metric := range metrics {
		if metric.DestVerb.Module == module && metric.DestVerb.Name == verb && metric.Name == metricName {
			switch count := metric.DataPoint.(type) {
			case dal.MetricCounter:
				return &pbconsole.MetricCounter{
					Value: count.Value,
				}
			default:
				return nil
			}
		}
	}
	return nil
}

func getHistogramMetric(metrics []dal.Metric, module string, verb string, metricName string) *pbconsole.MetricHistorgram {
	for _, metric := range metrics {
		if metric.DestVerb.Module == module && metric.DestVerb.Name == verb && metric.Name == metricName {
			switch histogram := metric.DataPoint.(type) {
			case dal.MetricHistogram:
				return &pbconsole.MetricHistorgram{
					Sum:    histogram.Sum,
					Count:  histogram.Count,
					Bucket: histogram.Bucket,
				}
			default:
				return nil
			}
		}
	}
	return nil
}
