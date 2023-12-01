package logger

import (
	"log/slog"
	"os"

	"github.com/asynkron/protoactor-go/actor"
)

// New is a logger factory
func New(system *actor.ActorSystem) *slog.Logger {
	level := new(slog.LevelVar)
	level.Set(slog.LevelInfo)
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})).
		With("lib", "Proto.Actor").
		With("system", system.ID)
}
