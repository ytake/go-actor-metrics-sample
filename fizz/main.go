package main

import (
	"context"
	"fmt"
	"log"

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

// FizzGrain / Virtual Actor
type FizzGrain struct {
	// persistence.Mixin
}

func (s *FizzGrain) Init(ctx cluster.GrainContext)      {}
func (s *FizzGrain) Terminate(ctx cluster.GrainContext) {}

func (s *FizzGrain) ReceiveDefault(ctx cluster.GrainContext) {
	fmt.Println(ctx.Message())
	switch ctx.Message().(type) {
	case *persistence.RequestSnapshot:
		// Handle snapshot request
	case *persistence.ReplayComplete:
		// Handle snapshot recovery
	}
}

func (s *FizzGrain) SayFizz(request *shared.FizzRequest, ctx cluster.GrainContext) (*shared.FizzResponse, error) {
	fmt.Printf("Received SayFizz with message '%v'\n", request.Message)
	sender := ctx.Sender()
	log.Printf("Received Ping call from sender. Address: %s. ID: %s.", sender.GetAddress(), sender.GetId())
	return &shared.FizzResponse{Message: "Fizz"}, nil
}

func main() {

	ctx := context.Background()
	exporter, err := metrics.NewOpenTelemetry("127.0.0.1:4318", "s").Exporter(ctx)
	if err != nil {
		panic(err)
	}

	system := actor.NewActorSystemWithConfig(
		actor.Configure(actor.WithMetricProviders(exporter)))
	provider, _ := zk.New([]string{"localhost:2181", "localhost:2182", "localhost:2183"})
	lookup := disthash.New()
	config := remote.Configure("localhost", 0)
	clusterConfig := cluster.Configure("fizzbuzz-cluster", provider, lookup, config,
		cluster.WithKinds(shared.NewFizzServiceKind(func() shared.FizzService {
			return &FizzGrain{}
		}, 0)))

	c := cluster.New(system, clusterConfig)
	c.StartMember()
	fmt.Print("\nBoot other nodes and press Enter\n")
	_, _ = console.ReadLine()
	c.Shutdown(true)
}
