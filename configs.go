package main

import (
	"fmt"
	"os"
	"strconv"
)

func generateSSLMode(env string) string {
	if env == "production" {
		return "require"
	}
	return "disable"
}

// GenerateConnectionString returns a connection string for a PostgreSQL database using environment variables
func GenerateConnectionString() string {
	// Get the database credentials from environment variables
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USERNAME")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbSSLMode := generateSSLMode(os.Getenv("ENV"))
	// Create the connection string
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", dbUser, dbPassword, dbHost, dbPort, dbName, dbSSLMode)

	return connStr
}

// GenerateJWTSecret returns a secret key for JWT token generation
func GenerateJWTSecret() string {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "secret"
	}
	return jwtSecret
}

// GenerateBCryptSalt returns the salt length for bcrypt hashing
func GenerateBCryptSalt() int {
	saltLength := os.Getenv("BCRYPT_SALT")
	if saltLength == "" {
		return 8
	}
	salt, err := strconv.Atoi(saltLength)
	if err != nil {
		return 8
	}
	return salt
}
