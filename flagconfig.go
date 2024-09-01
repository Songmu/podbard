package primcast

import (
	"context"
)

func withFlagConfig(ctx context.Context, cfg *flagConfig) context.Context {
	return context.WithValue(ctx, "flagConfig", cfg)
}

func getFlagConfig(ctx context.Context) *flagConfig {
	return ctx.Value("flagConfig").(*flagConfig)
}

type flagConfig struct {
	RootDir string
}
