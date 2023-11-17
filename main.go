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
	"github.com/asynkron/protoactor-go/stream"
	"github.com/ytake/go-actor-metrics-sample/clog"
	"github.com/ytake/go-actor-metrics-sample/metrics"
	"github.com/ytake/go-actor-metrics-sample/shared"
)

const rangeTo = 100

func main() {

	ctx := context.Background()
	exporter, err := metrics.NewNrOpenTelemetry(
		os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
		"actor-host",
		os.Getenv("NR_API_KEY")).Exporter(ctx)
	if err != nil {
		panic(err)
	}
	clog.SetLogLevel(plog.ErrorLevel)
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
	p := stream.NewTypedStream[*shared.FizzBuzzResponse](system)
	go func() {
		fizzbuzz := actor.PropsFromProducer(func() actor.Actor {
			return &FizzBuzz{
				system: system,
				pipe:   p.PID(),
			}
		})
		pid := system.Root.Spawn(fizzbuzz)
		for v := range [rangeTo]int64{} {
			system.Root.Send(pid, &shared.FizzBuzzRequest{Number: int64(v + 1)})
		}
	}()
	for range [rangeTo]int{} {
		fmt.Println(<-p.C())
	}
	_, _ = console.ReadLine()
}

type FizzBuzz struct {
	system *actor.ActorSystem
	pipe   *actor.PID
}

func (state *FizzBuzz) Receive(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *shared.FizzBuzzRequest:
		client := shared.GetFizzServiceGrainClient(
			cluster.GetCluster(state.system), "grain1")
		res, _ := client.SayFizzBuzz(&shared.FizzBuzzRequest{
			Number: msg.Number,
		})
		ctx.Send(state.pipe, res)
	}
}
