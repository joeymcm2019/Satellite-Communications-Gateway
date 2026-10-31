package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	HTTPPort string
	GRPCPort string
}

func Load() (Config, error) {
	c := Config{HTTPPort: env("GATEWAY_HTTP_PORT", "8080"), GRPCPort: env("GATEWAY_GRPC_PORT", "50051")}
	for name, value := range map[string]string{"HTTP port": c.HTTPPort, "gRPC port": c.GRPCPort} {
		port, err := strconv.Atoi(value)
		if err != nil || port < 1 || port > 65535 {
			return Config{}, fmt.Errorf("invalid %s: %q", name, value)
		}
	}
	return c, nil
}
func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
