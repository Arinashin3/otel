package config

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	sdkLog "go.opentelemetry.io/otel/sdk/log"
	sdkMetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

func (_cfg *Configuration) GenerateMeterProviders(ctx context.Context, serviceName string) map[string]*sdkMetric.MeterProvider {
	if _cfg.loaded == false {
		_cfg.logger.Error("configuration not loaded")
		return nil
	}
	if !_cfg.Server.Metrics.Enabled {
		_cfg.logger.Info("metrics disabled")
		return nil
	}
	var exp sdkMetric.Exporter
	var err error
	cfg := _cfg.Server.Metrics
	endpoint := fmt.Sprintf("%s%s", *cfg.Endpoint, *cfg.Api_path)
	// Exporter
	switch *cfg.Mode {
	case "http":
		if *cfg.Insecure {
			exp, err = otlpmetrichttp.New(ctx,
				otlpmetrichttp.WithEndpointURL(endpoint),
				otlpmetrichttp.WithInsecure(),
			)
		} else {
			exp, err = otlpmetrichttp.New(ctx,
				otlpmetrichttp.WithEndpointURL(endpoint),
			)
		}
	case "grpc":
		if *cfg.Insecure {
			exp, err = otlpmetricgrpc.New(ctx,
				otlpmetricgrpc.WithEndpointURL(endpoint),
				otlpmetricgrpc.WithInsecure(),
			)
		} else {
			exp, err = otlpmetricgrpc.New(ctx,
				otlpmetricgrpc.WithEndpointURL(endpoint),
			)
		}
	}

	if err != nil {
		_cfg.logger.Error("failed to initialize metrics exporter")
		return nil
	}
	var mps = make(map[string]*sdkMetric.MeterProvider)
	for _, client := range _cfg.Clients {
		var attrs []attribute.KeyValue
		for k, v := range client.Labels {
			attrs = append(attrs, attribute.String(k, v))
		}
		attrs = append(attrs, attribute.String("service.name", serviceName))
		mps[*client.Endpoint] = sdkMetric.NewMeterProvider(
			sdkMetric.WithResource(
				resource.NewSchemaless(attrs...),
			),
			sdkMetric.WithReader(
				sdkMetric.NewPeriodicReader(exp,
					sdkMetric.WithInterval(*client.Interval),
				),
			),
		)

	}
	return mps
}

func (_cfg *Configuration) GenerateLoggerProviders(ctx context.Context, serviceName string) map[string]*sdkLog.LoggerProvider {
	if _cfg.loaded == false {
		_cfg.logger.Error("configuration not loaded")
		return nil
	}
	if !_cfg.Server.Logs.Enabled {
		_cfg.logger.Error("logs disabled")
		return nil
	}

	cfg := _cfg.Server.Logs
	endpoint := fmt.Sprintf("%s%s", *cfg.Endpoint, *cfg.Api_path)

	var exp sdkLog.Exporter
	var err error
	switch *cfg.Mode {
	case "http":
		if *cfg.Insecure {
			exp, err = otlploghttp.New(ctx,
				otlploghttp.WithEndpointURL(endpoint),
				otlploghttp.WithInsecure(),
			)
		} else {
			exp, err = otlploghttp.New(ctx,
				otlploghttp.WithEndpointURL(endpoint),
			)
		}
	case "grpc":
		if *cfg.Insecure {
			exp, err = otlploggrpc.New(ctx,
				otlploggrpc.WithEndpointURL(endpoint),
				otlploggrpc.WithInsecure(),
			)
		} else {
			exp, err = otlploggrpc.New(ctx,
				otlploggrpc.WithEndpointURL(endpoint),
			)
		}
	}
	if err != nil {
		_cfg.logger.Error("failed to initialize logger provider")
		return nil
	}

	var lps = make(map[string]*sdkLog.LoggerProvider)
	for _, client := range _cfg.Clients {
		var attrs []attribute.KeyValue
		for k, v := range client.Labels {
			attrs = append(attrs, attribute.String(k, v))
		}
		attrs = append(attrs, attribute.String("service.name", serviceName))
		lps[*client.Endpoint] = sdkLog.NewLoggerProvider(
			sdkLog.WithResource(
				resource.NewSchemaless(attrs...),
			),
			sdkLog.WithProcessor(
				sdkLog.NewSimpleProcessor(exp),
			),
		)
	}

	return lps
}
