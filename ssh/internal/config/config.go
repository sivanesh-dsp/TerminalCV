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
	APIURL      string        // public résumé REST API, advertised in the Contact tab
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
//
// Note on the REST API: the SSH server reads shared/resume.json directly — the
// same file the API serves — so it needs no HTTP calls to stay in sync. APIURL
// is only advertised to visitors who want to consume the résumé programmatically.
func Load() Config {
	return Config{
		SSHAddr:     env("SSH_ADDR", ":2222"),
		HealthAddr:  env("HEALTH_ADDR", ":8081"),
		HostKeyPath: env("HOST_KEY_PATH", "data/host_ed25519"),
		ResumePath:  env("RESUME_PATH", ""),
		StatePath:   env("STATE_PATH", "data/state.json"),
		APIURL:      env("RESUME_API_URL", ""),
		IdleTimeout: envDur("IDLE_TIMEOUT", 5*time.Minute),
		MaxTimeout:  envDur("MAX_TIMEOUT", 60*time.Minute),
	}
}
