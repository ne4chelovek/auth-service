package tracing

import (
	"github.com/uber/jaeger-client-go/config"
	"go.uber.org/zap"
)

func Init(logger *zap.Logger, serviceName string) {
	cfg := config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 0.5,
		},
	}
	_, err := cfg.InitGlobalTracer(serviceName)
	if err != nil {
		logger.Fatal("Failed to init tracer", zap.Error(err))
	}
}
