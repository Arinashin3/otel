package config

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
)

func TestConfiguration(t *testing.T) {
	var logger *slog.Logger
	logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource:   true,
		Level:       slog.LevelDebug,
		ReplaceAttr: nil,
	}))
	cfg := NewConfiguration()
	err := cfg.LoadFile("testData/full-config.yaml", logger)
	ctx := context.Background()
	mp := cfg.GenerateMeterProviders(ctx, "test")
	lp := cfg.GenerateLoggerProviders(ctx, "test")

	fmt.Println(mp, lp, err)

}
