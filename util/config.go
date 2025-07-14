package util

import (
	"log"
	"os"
	"strconv"
	"strings"
)

// LoadEnvFile loads environment variables from a .env file
func LoadEnvFile(filename string) error {
	log.Printf("UTIL: Attempting to load .env file: %s", filename)
	
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// File doesn't exist, skip loading
		log.Printf("UTIL: .env file does not exist: %s", filename)
		return nil
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		log.Printf("UTIL: Error reading .env file: %v", err)
		return err
	}

	log.Printf("UTIL: Successfully read .env file, %d bytes", len(data))
	
	lines := strings.Split(string(data), "\n")
	loadedVars := 0
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'')) {
			value = value[1 : len(value)-1]
		}

		log.Printf("UTIL: Setting environment variable: %s=%s", key, value)
		os.Setenv(key, value)
		loadedVars++
	}

	log.Printf("UTIL: Loaded %d environment variables from .env file", loadedVars)
	return nil
}

// GetEnvString gets a string environment variable with a default value
func GetEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvInt gets an integer environment variable with a default value
func GetEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// GetEnvBool gets a boolean environment variable with a default value
func GetEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}
