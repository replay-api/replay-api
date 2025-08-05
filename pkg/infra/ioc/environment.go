package ioc

import (
	"os"
	"strconv"
	"strings"

	common "github.com/replay-api/replay-api/pkg/domain"
)

func EnvironmentConfig() (common.Config, error) {
	config := common.Config{
		Auth: common.AuthConfig{
			SteamConfig: common.SteamConfig{
				SteamKey:    os.Getenv("STEAM_KEY"),
				PublicKey:   os.Getenv("STEAM_PUB_KEY"),
				Certificate: os.Getenv("STEAM_CERT"),
				VHashSource: os.Getenv("STEAM_VHASH_SOURCE"),
			},
			BattleNetConfig: common.BattleNetConfig{
				BattleNetKey: os.Getenv("BATTLENET_KEY"),
			},
			GitHubConfig: common.GitHubConfig{
				GitHubKey: os.Getenv("GITHUB_KEY"),
			},
		},
		MongoDB: common.MongoDBConfig{
			URI:         os.Getenv("MONGO_URI"),
			PublicKey:   os.Getenv("MONGO_PUB_KEY"),
			Certificate: os.Getenv("MONGO_CERT"),
			DBName:      os.Getenv("MONGO_DB_NAME"),
		},
		CORS: loadCORSConfig(),
	}

	return config, nil
}

func loadCORSConfig() common.CORSConfig {
	corsOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
	if corsOrigins == "" {
		corsOrigins = "*" // Default to allow all origins
	}

	corsMethods := os.Getenv("CORS_ALLOWED_METHODS")
	if corsMethods == "" {
		corsMethods = "GET,POST,PUT,DELETE,OPTIONS" // Default methods
	}

	corsHeaders := os.Getenv("CORS_ALLOWED_HEADERS")
	if corsHeaders == "" {
		corsHeaders = "Content-Type,Authorization,X-Requested-With" // Default headers
	}

	allowCredentials := false
	if credStr := os.Getenv("CORS_ALLOW_CREDENTIALS"); credStr != "" {
		allowCredentials, _ = strconv.ParseBool(credStr)
	}

	maxAge := 86400 // 24 hours default
	if maxAgeStr := os.Getenv("CORS_MAX_AGE"); maxAgeStr != "" {
		if parsedMaxAge, err := strconv.Atoi(maxAgeStr); err == nil {
			maxAge = parsedMaxAge
		}
	}

	return common.CORSConfig{
		AllowedOrigins:   strings.Split(corsOrigins, ","),
		AllowedMethods:   strings.Split(corsMethods, ","),
		AllowedHeaders:   strings.Split(corsHeaders, ","),
		AllowCredentials: allowCredentials,
		MaxAge:           maxAge,
	}
}
