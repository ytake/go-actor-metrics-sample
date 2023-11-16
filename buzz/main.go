package main

import (
	"context"
	"os"

	console "github.com/asynkron/goconsole"
	"github.com/asynkron/protoactor-go/actor"
	"github.com/asynkron/protoactor-go/cluster"
	"github.com/asynkron/protoactor-go/cluster/clusterproviders/zk"
	"github.com/asynkron/protoactor-go/cluster/identitylookup/disthash"
	plog "github.com/asynkron/protoactor-go/log"
	"github.com/asynkron/protoactor-go/remote"
	"github.com/ytake/go-actor-metrics-sample/clog"
	"github.com/ytake/go-actor-metrics-sample/metrics"
	"github.com/ytake/go-actor-metrics-sample/shared"
)

// BuzzGrain / Virtual Actor
type BuzzGrain struct{}

func (s *BuzzGrain) Init(ctx cluster.GrainContext)           {}
func (s *BuzzGrain) Terminate(ctx cluster.GrainContext)      {}
func (s *BuzzGrain) ReceiveDefault(ctx cluster.GrainContext) {}

func (s *BuzzGrain) SayBuzz(request *shared.FizzBuzzRequest, ctx cluster.GrainContext) (*shared.FizzBuzzResponse, error) {
	response := &shared.FizzBuzzResponse{Message: request.Message}
	response.Number = request.Number
	if request.Number%5 == 0 {
		response.Message = response.Message + "Buzz"
	}
	return response, nil
}

func main() {
	ctx := context.Background()
	exporter, err := metrics.NewNrOpenTelemetry(
		os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
		"buzz",
		os.Getenv("NR_API_KEY")).Exporter(ctx)
	if err != nil {
		os.Exit(1)
	}
	clog.SetLogLevel(plog.ErrorLevel)
	system := actor.NewActorSystemWithConfig(
		actor.Configure(actor.WithMetricProviders(exporter)))
	provider, _ := zk.New([]string{"localhost:2181", "localhost:2182", "localhost:2183"})
	lookup := disthash.New()
	config := remote.Configure("localhost", 0)
	clusterConfig := cluster.Configure("fizzbuzz-cluster", provider, lookup, config,
		cluster.WithKinds(shared.NewBuzzServiceKind(func() shared.BuzzService {
			return &BuzzGrain{}
		}, 100)))
	c := cluster.New(system, clusterConfig)
	c.StartMember()
	_, _ = console.ReadLine()
	c.Shutdown(true)
}
