package config

import "github.com/spf13/viper"

var env *EnvConfig

type EnvConfig struct {
	DBUrl           string `mapstructure:"DATABASE_URL"`
	DBUrl_Migration string `mapstructure:"MIGRATION_DATABASE_URL"`
	PORT            string `mapstructure:"PORT"`
	ENV             string `mapstructure:"ENVIRONMENT"`
	MIGRATIONS_PATH string `mapstructure:"MIGRATIONS_PATH"`
	JWT_SECRET      string `mapstructure:"JWT_SECRET"`
}

func GetEnvConfig() *EnvConfig {
	return env
}

func LoadEnv(path string) (*EnvConfig, error) {
	viper.SetConfigName("app_config")
	viper.SetConfigType("env")
	viper.AddConfigPath(path)
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	// Explicitly bind the OS environment variables to the keys expected by mapstructure
	viper.BindEnv("DATABASE_URL")
	viper.BindEnv("MIGRATION_DATABASE_URL")
	viper.BindEnv("PORT")
	viper.BindEnv("ENVIRONMENT")
	viper.BindEnv("MIGRATIONS_PATH")
	viper.BindEnv("JWT_SECRET")

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
