package podbard

import (
	"context"
)

type ctxkey string

const (
	flagConfigKey ctxkey = "flagConfig"
)

func withFlagConfig(ctx context.Context, cfg *flagConfig) context.Context {
	return context.WithValue(ctx, flagConfigKey, cfg)
}

func getFlagConfig(ctx context.Context) *flagConfig {
	return ctx.Value(flagConfigKey).(*flagConfig)
}

type flagConfig struct {
	RootDir string
}
