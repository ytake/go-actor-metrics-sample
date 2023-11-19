package main

import (
	"context"
	"fmt"
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

// FizzGrain / Virtual Actor
type FizzGrain struct{}

func (s *FizzGrain) Init(ctx cluster.GrainContext)      {}
func (s *FizzGrain) Terminate(ctx cluster.GrainContext) {}

// ReceiveDefault is the default handler.
// gRPCではなく、通常のメッセージを受信する場合はこちらを利用します
func (s *FizzGrain) ReceiveDefault(ctx cluster.GrainContext) {
	switch msg := ctx.Message().(type) {
	case *shared.FizzBuzzRequest:
		fmt.Println(msg)
	}
}

func (s *FizzGrain) SayFizzBuzz(request *shared.FizzBuzzRequest, ctx cluster.GrainContext) (*shared.FizzBuzzResponse, error) {
	r := &shared.FizzBuzzRequest{Message: ""}
	r.Number = request.Number
	if request.Number%3 == 0 {
		r.Message = "Fizz"
	}
	client := shared.GetBuzzServiceGrainClient(cluster.GetCluster(ctx.ActorSystem()), "grain2")
	return client.SayBuzz(r)
}

func main() {
	ctx := context.Background()
	// docker環境に送信する場合は下記のように設定します
	// exporter, err := metrics.NewOpenTelemetry("127.0.0.1:4318", "actor-host").Exporter(ctx)
	// NewRelicに送信する場合は下記のように設定します
	exporter, err := metrics.NewNrOpenTelemetry(
		os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
		os.Getenv("NR_API_KEY"),
		"actor-host").Exporter(ctx)
	if err != nil {
		os.Exit(1)
	}
	system := actor.NewActorSystemWithConfig(actor.Configure(actor.WithMetricProviders(exporter)))
	clog.SetLogLevel(plog.ErrorLevel)
	provider, _ := zk.New([]string{"localhost:2181", "localhost:2182", "localhost:2183"})
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
