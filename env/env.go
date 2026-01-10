package env

import (
	"sync"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

// Build-time variables (injected via ldflags)
var (
	BuildWebhookURL string // Injected at build time via -ldflags "-X github.com/verbeux-ai/whatsmiau/env.BuildWebhookURL=..."
)

type E struct {
	Port           string `env:"PORT" envDefault:"8080"`
	DebugMode      bool   `env:"DEBUG_MODE" envDefault:"false"`
	DebugWhatsmeow bool   `env:"DEBUG_WHATSMEOW" envDefault:"false"`

	RedisURL      string `env:"REDIS_URL" envDefault:"localhost:6379"`
	RedisPassword string `env:"REDIS_PASSWORD"`
	RedisTLS      bool   `env:"REDIS_TLS" envDefault:"false"`

	ApiKey    string `env:"API_KEY" envDefault:""`
	DBDialect string `env:"DIALECT_DB" envDefault:"sqlite3"`                   // sqlite3 or postgres
	DBURL     string `env:"DB_URL" envDefault:"file:data.db?_foreign_keys=on"` // "postgres://<user>:<pass>@<host>:<port>/<DB>?sslmode=disable

	GCSEnabled bool   `env:"GCS_ENABLED" envDefault:"false"`
	GCSBucket  string `env:"GCS_BUCKET" envDefault:"whatsmiau"`
	GCSURL     string `env:"GCS_URL" envDefault:"https://storage.googleapis.com"`

	GCL          string `json:"GCL_APP_NAME" envDefault:"whatsmiau-br-1"`
	GCLEnabled   bool   `json:"GCL_ENABLED" envDefault:"false"`
	GCLProjectID string `json:"GCL_PROJECT_ID"`

	EmitterBufferSize    int `env:"EMITTER_BUFFER_SIZE" envDefault:"2048"`
	HandlerSemaphoreSize int `env:"HANDLER_SEMAPHORE_SIZE" envDefault:"512"`

	ProxyAddresses []string `env:"PROXY_ADDRESSES" envDefault:""`      // random choices proxies ex: <SOCKS5|HTTP|HTTPS>://<username>:<password>@<host>:<port>
	ProxyStrategy  string   `env:"PROXY_STRATEGY" envDefault:"RANDOM"` // todo: implement BALANCED
	ProxyNoMedia   bool     `env:"PROXY_NO_MEDIA" envDefault:"false"`

	WebhookURL string `env:"WEBHOOK_URL" envDefault:""` // URL to send system metrics
}

var (
	Env E
	webhookURLMutex sync.RWMutex
	dynamicWebhookURL string // Can be updated via API
)

// GetWebhookURL returns the webhook URL with priority:
// 1. Dynamic URL (set via API)
// 2. Environment variable
// 3. Build-time variable
func GetWebhookURL() string {
	webhookURLMutex.RLock()
	defer webhookURLMutex.RUnlock()

	if dynamicWebhookURL != "" {
		return dynamicWebhookURL
	}
	if Env.WebhookURL != "" {
		return Env.WebhookURL
	}
	return BuildWebhookURL
}

// SetWebhookURL sets the webhook URL dynamically (thread-safe)
func SetWebhookURL(url string) {
	webhookURLMutex.Lock()
	defer webhookURLMutex.Unlock()
	dynamicWebhookURL = url
}

func Load() error {
	_ = godotenv.Load(".env")
	err := env.Parse(&Env)

	// If build-time webhook URL is set and no env/webhook URL, use build-time
	if BuildWebhookURL != "" && Env.WebhookURL == "" && dynamicWebhookURL == "" {
		dynamicWebhookURL = BuildWebhookURL
	}

	return err
}
