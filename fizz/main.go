package main

import (
	"context"
	"fmt"
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

// FizzGrain / Virtual Actor
type FizzGrain struct{}

func (s *FizzGrain) Init(ctx cluster.GrainContext)      {}
func (s *FizzGrain) Terminate(ctx cluster.GrainContext) {}

// ReceiveDefault is the default handler.
// gRPCではなく、通常のメッセージを受信する場合はこちらを利用します
func (s *FizzGrain) ReceiveDefault(ctx cluster.GrainContext) {
	switch msg := ctx.Message().(type) {
	case *shared.FizzRequest:
		fmt.Println(msg)
	}
}

func (s *FizzGrain) SayFizz(request *shared.FizzRequest, ctx cluster.GrainContext) (*shared.FizzResponse, error) {
	r := &shared.FizzResponse{Message: ""}
	r.Number = request.Number
	if request.Number%3 == 0 {
		r.Message = "Fizz"
	}
	return r, nil
}

func main() {
	ctx := context.Background()
	// docker環境に送信する場合は下記のように設定します
	// meterProvider, err := metrics.NewOpenTelemetry("127.0.0.1:4318", "fizz").Exporter(ctx)
	// NewRelicに送信する場合は下記のように設定します
	meterProvider, err := metrics.NewNrOpenTelemetry(
		os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
		os.Getenv("NR_API_KEY"),
		"fizz").Exporter(ctx)
	if err != nil {
		os.Exit(1)
	}
	system := actor.NewActorSystemWithConfig(
		actor.Configure(
			actor.WithMetricProviders(meterProvider),
			actor.WithLoggerFactory(logger.New)))
	provider, _ := consul.New()
	lookup := disthash.New()
	config := remote.Configure("localhost", 0)
	clusterConfig := cluster.Configure("fizzbuzz-cluster", provider, lookup, config,
		cluster.WithKinds(shared.NewFizzServiceKind(func() shared.FizzService {
			return &FizzGrain{}
		}, 100)))

	c := cluster.New(system, clusterConfig)
	c.StartMember()

	_, _ = console.ReadLine()
	c.Shutdown(true)
}
