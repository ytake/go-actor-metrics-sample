package clog

import (
	"github.com/asynkron/protoactor-go/actor"
	"github.com/asynkron/protoactor-go/cluster"
	"github.com/asynkron/protoactor-go/cluster/clusterproviders/zk"
	"github.com/asynkron/protoactor-go/cluster/identitylookup/disthash"
	plog "github.com/asynkron/protoactor-go/log"
	"github.com/asynkron/protoactor-go/remote"
)

func SetLogLevel(level plog.Level) {
	actor.SetLogLevel(level)
	cluster.SetLogLevel(level)
	remote.SetLogLevel(level)
	zk.SetLogLevel(level)
	disthash.SetLogLevel(level)
}
