package main

import (
	"context"
	"fmt"

	console "github.com/asynkron/goconsole"
	"github.com/asynkron/protoactor-go/actor"
	"github.com/asynkron/protoactor-go/cluster"
	"github.com/asynkron/protoactor-go/cluster/clusterproviders/zk"
	"github.com/asynkron/protoactor-go/cluster/identitylookup/disthash"
	"github.com/asynkron/protoactor-go/persistence"
	"github.com/asynkron/protoactor-go/remote"
	"github.com/ytake/go-actor-metrics-sample/metrics"
	"github.com/ytake/go-actor-metrics-sample/shared"
)

// Fizz Grain / aka Virtual Actor
type Fizz struct {
	persistence.Mixin
}

func (s *Fizz) Init(ctx cluster.GrainContext)      {}
func (s *Fizz) Terminate(ctx cluster.GrainContext) {}

func (s *Fizz) ReceiveDefault(ctx cluster.GrainContext) {
	switch ctx.Message().(type) {
	case *persistence.RequestSnapshot:
		// Handle snapshot request
	case *persistence.ReplayComplete:
		// Handle snapshot recovery
	}
}

func (s *Fizz) SayFizz(request *shared.FizzRequest, ctx cluster.GrainContext) (*shared.FizzResponse, error) {
	return &shared.FizzResponse{Message: "Fizz"}, nil
}

func main() {

	ctx := context.Background()
	exporter, err := metrics.NewOpenTelemetry("e", "s").Exporter(ctx)
	if err != nil {
		panic(err)
	}

	system := actor.NewActorSystemWithConfig(
		actor.Configure(actor.WithMetricProviders(exporter)))
	provider, _ := zk.New([]string{"localhost:2181", "localhost:2182", "localhost:2183"})
	lookup := disthash.New()
	config := remote.Configure("localhost", 0)
	clusterConfig := cluster.Configure("fizzbuzz-cluster", provider, lookup, config)
	c := cluster.New(system, clusterConfig)
	c.StartMember()
	fmt.Print("\nBoot other nodes and press Enter\n")
	_, _ = console.ReadLine()
}
