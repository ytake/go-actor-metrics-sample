package main

import (
	"context"
	"os"

	console "github.com/asynkron/goconsole"
	"github.com/asynkron/protoactor-go/actor"
	"github.com/asynkron/protoactor-go/cluster"
	"github.com/asynkron/protoactor-go/cluster/clusterproviders/consul"
	"github.com/asynkron/protoactor-go/cluster/identitylookup/disthash"
	"github.com/asynkron/protoactor-go/remote"
	"github.com/ytake/go-actor-metrics-sample/logger"
	"github.com/ytake/go-actor-metrics-sample/metrics"
	"github.com/ytake/go-actor-metrics-sample/shared"
)

// BuzzGrain / Virtual Actor
type BuzzGrain struct{}

func (s *BuzzGrain) Init(_ cluster.GrainContext)           {}
func (s *BuzzGrain) Terminate(_ cluster.GrainContext)      {}
func (s *BuzzGrain) ReceiveDefault(_ cluster.GrainContext) {}

func (s *BuzzGrain) SayBuzz(request *shared.BuzzRequest, _ cluster.GrainContext) (*shared.BuzzResponse, error) {
	response := &shared.BuzzResponse{Message: request.Message}
	response.Number = request.Number
	if request.Number%5 == 0 {
		response.Message = response.Message + "Buzz"
	}
	return response, nil
}

func main() {
	ctx := context.Background()
	// docker環境に送信する場合は下記のように設定します
	// exporter, err := metrics.NewOpenTelemetry("127.0.0.1:4318", "actor-host").Exporter(ctx)
	// NewRelicに送信する場合は下記のように設定します
	exporter, err := metrics.NewNrOpenTelemetry(
		os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
		os.Getenv("NR_API_KEY"),
		"buzz").Exporter(ctx)
	if err != nil {
		os.Exit(1)
	}
	system := actor.NewActorSystemWithConfig(
		actor.Configure(
			actor.WithMetricProviders(exporter),
			actor.WithLoggerFactory(logger.New)))
	provider, _ := consul.New()
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
