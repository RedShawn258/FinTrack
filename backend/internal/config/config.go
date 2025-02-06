package config

import "os"

func LoadConfig() {
	os.Setenv("DB_HOST", "localhost") // Example config
}
