package main

import (
	"context"
	"fmt"

	console "github.com/asynkron/goconsole"
	"github.com/asynkron/protoactor-go/actor"
	"github.com/asynkron/protoactor-go/cluster"
	"github.com/asynkron/protoactor-go/cluster/clusterproviders/zk"
	"github.com/asynkron/protoactor-go/cluster/identitylookup/disthash"
	"github.com/asynkron/protoactor-go/remote"
	"github.com/ytake/go-actor-metrics-sample/metrics"
	"github.com/ytake/go-actor-metrics-sample/shared"
)

func main() {

	ctx := context.Background()
	exporter, err := metrics.NewOpenTelemetry("127.0.0.1:4318", "host").Exporter(ctx)
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
	defer c.Shutdown(false)

	fmt.Print("\nBoot other nodes and press Enter\n")
	_, _ = console.ReadLine()

	fizzbuzz := actor.PropsFromProducer(func() actor.Actor {
		return &FizzBuzz{
			system: system,
		}
	})
	pid := system.Root.Spawn(fizzbuzz)
	system.Root.Send(pid, &shared.FizzRequest{
		Message: "hello",
	})
	console.ReadLine()
}

type FizzBuzz struct {
	system *actor.ActorSystem
}

func (state *FizzBuzz) Receive(ctx actor.Context) {
	switch ctx.Message().(type) {
	case *shared.FizzRequest:
		client := shared.GetFizzServiceGrainClient(
			cluster.GetCluster(state.system), "grain1")
		res, _ := client.SayFizz(&shared.FizzRequest{
			Message: "hello",
		})
		fmt.Printf("Response1: %v\n", res)
		fmt.Println()
	case *shared.FizzResponse:
		fmt.Printf("Response2: %v\n", ctx.Message())
	}
}
