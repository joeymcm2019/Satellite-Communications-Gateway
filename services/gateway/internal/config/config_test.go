package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	t.Setenv("GATEWAY_HTTP_PORT", "")
	t.Setenv("GATEWAY_GRPC_PORT", "")
	cfg, err := Load()
	if err != nil || cfg.HTTPPort != "18080" || cfg.GRPCPort != "50051" {
		t.Fatalf("unexpected config: %+v, %v", cfg, err)
	}
}

func TestRejectInvalidPort(t *testing.T) {
	for _, value := range []string{"0", "65536", "abc"} {
		t.Run(value, func(t *testing.T) {
			t.Setenv("GATEWAY_HTTP_PORT", value)
			if _, err := Load(); err == nil {
				t.Fatal("expected invalid port error")
			}
		})
	}
}
