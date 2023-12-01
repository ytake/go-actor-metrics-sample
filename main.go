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
	"github.com/asynkron/protoactor-go/stream"
	"github.com/ytake/go-actor-metrics-sample/logger"
	"github.com/ytake/go-actor-metrics-sample/metrics"
	"github.com/ytake/go-actor-metrics-sample/shared"
)

const rangeTo = 100

type FizzBuzz struct {
	system *actor.ActorSystem
	pipe   *actor.PID
}

type Say struct {
	Number int64
}

type Response struct {
	Number  int64
	Message string
}

func (state *FizzBuzz) Receive(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *Say:
		fc := shared.GetFizzServiceGrainClient(
			cluster.GetCluster(state.system), "grain1")
		res, _ := fc.SayFizz(&shared.FizzRequest{
			Number: msg.Number,
		})
		nb := shared.GetBuzzServiceGrainClient(
			cluster.GetCluster(state.system), "grain1")
		buzz, _ := nb.SayBuzz(&shared.BuzzRequest{
			Number:  msg.Number,
			Message: res.Message,
		})
		ctx.Send(state.pipe, &Response{Number: msg.Number, Message: buzz.Message})
	}
}

func main() {
	ctx := context.Background()
	// docker環境に送信する場合は下記のように設定します
	// meterProvider, err := metrics.NewOpenTelemetry("127.0.0.1:4318", "actor-host").Exporter(ctx)
	// NewRelicに送信する場合は下記のように設定します
	meterProvider, err := metrics.NewNrOpenTelemetry(
		os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
		os.Getenv("NR_API_KEY"),
		"actor-host").Exporter(ctx)
	if err != nil {
		os.Exit(1)
	}
	system := actor.NewActorSystemWithConfig(
		actor.Configure(
			actor.WithMetricProviders(meterProvider),
			actor.WithLoggerFactory(logger.New)))
	provider, _ := consul.New()
	config := remote.Configure("localhost", 0)
	clusterConfig := cluster.Configure("fizzbuzz-cluster", provider, disthash.New(), config)
	c := cluster.New(system, clusterConfig)
	c.StartMember()
	defer c.Shutdown(false)
	fmt.Print("\nBoot other nodes and press Enter\n")
	_, _ = console.ReadLine()

	p := stream.NewTypedStream[*Response](system)
	go func() {
		fizzbuzz := actor.PropsFromProducer(func() actor.Actor {
			return &FizzBuzz{
				system: system,
				pipe:   p.PID(),
			}
		})
		pid := system.Root.Spawn(fizzbuzz)
		for v := range [rangeTo]int64{} {
			// gRPC を介してメッセージを送信します
			system.Root.Send(pid, &Say{Number: int64(v + 1)})
			// 標準的なメッセージを送信する場合は下記のようにPIDを指定し、利用できます
			// クラスタの場合は直接クラスタのメンバーのPIDを指定します
			// system.Root.Send(c.Get("grain1", "FizzService"),
			//	&shared.FizzBuzzRequest{Number: int64(v + 1)})
		}
	}()
	for range [rangeTo]int{} {
		fmt.Println(<-p.C())
	}
	_, _ = console.ReadLine()
}
