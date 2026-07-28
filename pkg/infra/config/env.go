package config

import (
	"strings"

	"github.com/spf13/viper"
)

var env *EnvConfig

type EnvConfig struct {
	DBType          string `mapstructure:"DATABASE_TYPE"`
	DBUrl           string `mapstructure:"DATABASE_URL"`
	DBUrl_Migration string `mapstructure:"MIGRATION_DATABASE_URL"`
	PORT            string `mapstructure:"PORT"`
	BASE_PATH       string `mapstructure:"BASE_PATH"`
	ENV             string `mapstructure:"ENVIRONMENT"`
	MIGRATIONS_PATH string `mapstructure:"MIGRATIONS_PATH"`
	JWT_SECRET      string `mapstructure:"JWT_SECRET"`
	MAIL_ENABLED    bool   `mapstructure:"MAIL_ENABLED"`
	MAIL_HOST       string `mapstructure:"MAIL_HOST"`
	MAIL_PORT       string `mapstructure:"MAIL_PORT"`
	MAIL_USER       string `mapstructure:"MAIL_USER"`
	MAIL_PASSWORD   string `mapstructure:"MAIL_PASSWORD"`
	MAIL_FROM       string `mapstructure:"MAIL_FROM"`
	MAIL_SECURE     bool   `mapstructure:"MAIL_SECURE"`
}

func GetEnvConfig() *EnvConfig {
	return env
}

func SetEnvConfigForTest(e *EnvConfig) {
	env = e
}

func (e *EnvConfig) GetDatabaseType() string {
	if e == nil {
		return "sqlite"
	}
	if e.DBType != "" {
		return strings.ToLower(e.DBType)
	}
	if strings.HasPrefix(strings.ToLower(e.DBUrl), "postgres") {
		return "postgres"
	}
	return "sqlite"
}

func LoadEnv(path string) (*EnvConfig, error) {
	viper.SetConfigName("app_config")
	viper.SetConfigType("env")
	viper.AddConfigPath(path)
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	// Explicitly bind the OS environment variables to the keys expected by mapstructure
	viper.BindEnv("DATABASE_TYPE")
	viper.BindEnv("DATABASE_URL")
	viper.BindEnv("MIGRATION_DATABASE_URL")
	viper.BindEnv("PORT")
	viper.BindEnv("BASE_PATH")
	viper.BindEnv("ENVIRONMENT")
	viper.BindEnv("MIGRATIONS_PATH")
	viper.BindEnv("JWT_SECRET")
	viper.BindEnv("MAIL_ENABLED")
	viper.BindEnv("MAIL_HOST")
	viper.BindEnv("MAIL_PORT")
	viper.BindEnv("MAIL_USER")
	viper.BindEnv("MAIL_PASSWORD")
	viper.BindEnv("MAIL_FROM")
	viper.BindEnv("MAIL_SECURE")

	err := viper.ReadInConfig()
	if err != nil {
		// Ignore error since the .env file might not exist,
		// and we will rely on OS environment variables.
	}

	err = viper.Unmarshal(&env)
	if err != nil {
		panic(err)
	}

	return env, nil
}
