package samsara

import (
	"bufio"
	"os"
	"strings"
)

type Config struct {
	DBPath string
}

// LoadConfig carga las variables de configuración desde el archivo .env
func LoadConfig(envPath string) (*Config, error) {
	config := &Config{
		DBPath: "./data", // Valor por defecto
	}

	file, err := os.Open(envPath)
	if err != nil {
		// Si no existe el archivo .env, retorna la configuración con defaults
		return config, nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Ignorar líneas vacías y comentarios
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "DB_PATH":
			config.DBPath = value
		}
	}

	return config, scanner.Err()
}
