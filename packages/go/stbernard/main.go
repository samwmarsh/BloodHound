package main

import (
	"os"

	"github.com/specterops/bloodhound/packages/go/stbernard/command"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
	"golang.org/x/exp/slog"
)

func main() {
	setLogger()

	if cmd, err := command.ParseCLI(); err != nil {
		slog.Error("Failed to parse CLI", "error", err)
		os.Exit(1)
	} else if err := cmd.Run(); err != nil {
		slog.Error("Failed to run command", "command", cmd.Name(), "error", err)
		os.Exit(1)
	} else {
		slog.Info("Command completed successfully", "command", cmd.Name())
	}
}

func setLogger() {
	var logLevel slog.Level

	envLvl := os.Getenv(environment.LogLevel.Env())

	switch envLvl {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel})))
}
