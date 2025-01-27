package main

import (
	"log/slog"
	"os"

	"github.com/timofurrer/influss/cmd"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(log)

	cmd := cmd.NewCommand(log)
	cmd.Parse()
	cmd.Run()
}
