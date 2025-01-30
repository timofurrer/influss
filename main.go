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
	err := cmd.Parse()
	if err != nil {
		log.Error("failed to parse command line", slog.Any("error", err))
		return
	}
	cmd.Run()
}
