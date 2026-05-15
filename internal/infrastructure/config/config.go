package config

import (
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	HTTP     HTTP
	DB       DB
	Redis    Redis
	Scraper  Scraper
	Coverage Coverage
	External External
	Log      Log
}

type Redis struct {
	// Either set REDIS_URL (e.g. rediss://default:pass@host:6380/0 — what
	// Railway/Upstash hand you) and leave the rest blank, or set REDIS_ADDR
	// + optional REDIS_PASSWORD/USERNAME for plain TCP. REDIS_URL wins if both.
	URL      string        `env:"REDIS_URL"`
	Addr     string        `env:"REDIS_ADDR"`
	Username string        `env:"REDIS_USERNAME"`
	Password string        `env:"REDIS_PASSWORD"`
	DB       int           `env:"REDIS_DB" envDefault:"0"`
	TTL      time.Duration `env:"REDIS_TTL" envDefault:"10m"`
}

type HTTP struct {
	Addr           string        `env:"HTTP_ADDR" envDefault:":8090"`
	ReadTimeout    time.Duration `env:"HTTP_READ_TIMEOUT" envDefault:"15s"`
	WriteTimeout   time.Duration `env:"HTTP_WRITE_TIMEOUT" envDefault:"30s"`
	AllowedOrigins []string      `env:"CORS_ALLOWED_ORIGINS" envSeparator:"," envDefault:"http://localhost:5173"`
}

type DB struct {
	URL      string `env:"DATABASE_URL,required"`
	MaxConns int32  `env:"DATABASE_MAX_CONNS" envDefault:"10"`
}

type Scraper struct {
	Interval   time.Duration `env:"SCRAPER_INTERVAL" envDefault:"5m"`
	City       string        `env:"SCRAPER_CITY" envDefault:"Kraków"`
	Country    string        `env:"SCRAPER_COUNTRY" envDefault:"PL"`
	PageSize   int           `env:"SCRAPER_PAGE_SIZE" envDefault:"100"`
	RatePerSec int           `env:"SCRAPER_RATE_PER_SECOND" envDefault:"1"`
}

type Coverage struct {
	PrecomputeInterval        time.Duration `env:"COVERAGE_PRECOMPUTE_INTERVAL" envDefault:"30m"`
	PrecomputeProvinceLimit   int           `env:"COVERAGE_PRECOMPUTE_PROVINCE_LIMIT" envDefault:"16"`
	// Must match the values the frontend actually requests — the cache key
	// includes cellMeters and recommendation limit, so a mismatch means the
	// warm populates rows that user requests never hit.
	DefaultCellMeters         int `env:"COVERAGE_DEFAULT_CELL_METERS" envDefault:"1500"`
	DefaultRecommendationsTop int `env:"COVERAGE_DEFAULT_RECOMMENDATIONS_TOP" envDefault:"50"`
	MinProvincePoints         int           `env:"COVERAGE_MIN_PROVINCE_POINTS" envDefault:"50"`
}

type External struct {
	InpostBaseURL string `env:"INPOST_BASE_URL" envDefault:"https://api-global-points.easypack24.net"`
}

type Log struct {
	Level  string `env:"LOG_LEVEL" envDefault:"info"`
	Format string `env:"LOG_FORMAT" envDefault:"json"`
}

func Load() (Config, error) {
	_ = godotenv.Load()
	var c Config
	if err := env.Parse(&c); err != nil {
		return c, err
	}
	return c, nil
}
