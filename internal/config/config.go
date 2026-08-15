package config

import (
	"log"
	"os"
	"strconv"
)
import "log/slog"
import "github.com/joho/godotenv"
import "github.com/caarlos0/env/v11"

type Config struct {
	Port          int    `env:"SERVER_PORT" envDefault:"8080"`
	PublicDomain  string `env:"PUBLIC_DOMAIN" envDefault:"http://localhost"`
	Secure        bool   `env:"SERVER_SECURE" envDefault:"true"`
	LinkSecretKey string `env:"LINK_SECRET_KEY" envRequired:"true"`
	LogDebug      bool   `env:"LOG_DEBUG" envDefault:"false"`
	AdminEmail    string `env:"ADMIN_EMAIL" envRequired:"true"`
	Postgres      Postgres
	Supabase      Supabase
	Auth          Auth
	SendGrid      SendGrid
	Resend        Resend
	Areal         Areal
	LockerService LockerService
}

type LockerService struct {
	LockId        string `env:"LOCKER_LOCK_ID" envRequired:"true"`
	ClientId      string `env:"LOCKER_CLIENT_ID" envRequired:"true"`
	RefreshToken  string `env:"LOCKER_REFRESH_TOKEN" envRequired:"true"`
	ClientSecret  string `env:"LOCKER_CLIENT_SECRET" envRequired:"true"`
	TokenLifespan int    `env:"LOCKER_TOKEN_LIFESPAN" envRequired:"true"` // days
	RefreshWindow int    `env:"LOCKER_REFRESH_WINDOW" envRequired:"true"` // days
	AccessToken   string `env:"LOCKER_ACCESS_TOKEN"`
	BackupCode    string `env:"LOCKER_BACKUP_CODE"`
}

type Areal struct {
	CabinetLockCode string `env:"AREAL_CABINET_LOCK_CODE"`
}

type Supabase struct {
	DatabaseURL string `env:"SUPABASE_DATABASE_URL"`
}

type Postgres struct {
	Host     string `env:"POSTGRES_HOST" envDefault:"localhost"`
	Port     int    `env:"POSTGRES_PORT" envDefault:"5432"`
	User     string `env:"POSTGRES_USER" envDefault:"user"`
	Password string `env:"POSTGRES_PASSWORD" envDefault:"user"`
	Db       string `env:"POSTGRES_DB" envDefault:"snpb"`
}

type SendGrid struct {
	APIKey      string `env:"SENDGRID_API_KEY" envRequired:"true"`
	FromEmail   string `env:"SENDGRID_FROM_EMAIL" envRequired:"true"`
	FromName    string `env:"SENDGRID_FROM_NAME" envDefault:"Newsletter Platform"`
	UseEURegion bool   `env:"SENDGRID_USE_EU_REGION" envDefault:"false"`
}

type Resend struct {
	APIKey string `env:"RESEND_API_KEY" envRequired:"true"`
}

type Auth struct {
	JwtSecretKey                  string `env:"JWT_SECRET_KEY" envRequired:"true"`
	BcryptPwdCost                 int    `env:"BCRYPT_PWD_COST" envDefault:"10"`
	BcryptRefreshTokenCost        int    `env:"BCRYPT_REFRESH_TOKEN_COST" envDefault:"10"`
	RefreshTokenLifespan          int    `env:"REFRESH_TOKEN_LIFESPAN" envDefault:"10"`          // days
	JwtLifespan                   int    `env:"JWT_LIFESPAN" envDefault:"15"`                    // minutes
	RefreshTokenRotationThreshold int    `env:"REFRESH_TOKEN_ROTATION_THRESHOLD" envDefault:"5"` // days before expiration
}

// recalculate all values, so everything in program is in seconds
func (c *Config) recalculate() {
	c.LockerService.TokenLifespan = c.LockerService.TokenLifespan * 60 * 60 * 24

	c.Auth.JwtLifespan = c.Auth.JwtLifespan * 60
	c.Auth.RefreshTokenLifespan = c.Auth.RefreshTokenLifespan * 60 * 60 * 24
	c.Auth.RefreshTokenRotationThreshold = c.Auth.RefreshTokenRotationThreshold * 60 * 60 * 24
}

func NewConfig() (cfg Config) {
	err := godotenv.Load(".env")
	if err != nil {
		slog.Warn("Error loading ..env file", "err", err)
	}

	err = env.Parse(&cfg)
	if err != nil {
		log.Fatal("Error parsing .env variables", "err", err)
	}

	if portStr := os.Getenv("PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			cfg.Port = port
		}
	}

	if railwayDomain := os.Getenv("RAILWAY_PUBLIC_DOMAIN"); railwayDomain != "" {
		cfg.PublicDomain = "https://" + railwayDomain
	}

	cfg.recalculate()

	return
}
