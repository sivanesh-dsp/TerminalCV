// Package config loads server configuration from environment variables,
// applying sensible defaults so the server runs with zero configuration.
package config

import (
	"os"
	"time"
)

type Config struct {
	SSHAddr     string        // listen address for SSH, e.g. ":2222"
	HealthAddr  string        // listen address for the HTTP health endpoint
	HostKeyPath string        // path to the persisted ed25519 host key
	ResumePath  string        // explicit path to shared/resume.json ("" = auto-discover)
	StatePath   string        // path to the persisted visitor-state JSON
	IdleTimeout time.Duration // disconnect idle sessions after this
	MaxTimeout  time.Duration // hard cap on session length
}

func env(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func envDur(key string, def time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

// Load reads configuration from the environment.
func Load() Config {
	return Config{
		SSHAddr:     env("SSH_ADDR", ":2222"),
		HealthAddr:  env("HEALTH_ADDR", ":8081"),
		HostKeyPath: env("HOST_KEY_PATH", "data/host_ed25519"),
		ResumePath:  env("RESUME_PATH", ""),
		StatePath:   env("STATE_PATH", "data/state.json"),
		IdleTimeout: envDur("IDLE_TIMEOUT", 5*time.Minute),
		MaxTimeout:  envDur("MAX_TIMEOUT", 60*time.Minute),
	}
}
