package ioc

import (
	"fmt"
	"net/url"
	"os"

	common "github.com/replay-api/replay-api/pkg/domain"
)

// buildMongoURI constructs a MongoDB connection URI with credentials if provided
func buildMongoURI() string {
	// First check if a full URI is provided
	uri := os.Getenv("MONGO_URI")

	// If MONGODB_USER and MONGODB_PASSWORD are provided, inject them into the URI
	user := os.Getenv("MONGODB_USER")
	password := os.Getenv("MONGODB_PASSWORD")

	if user != "" && password != "" {
		// Parse the existing URI
		parsed, err := url.Parse(uri)
		if err == nil && parsed.User == nil {
			// No credentials in the URI, add them
			parsed.User = url.UserPassword(user, password)
			// Add authSource=admin for MongoDB with authentication
			q := parsed.Query()
			if q.Get("authSource") == "" {
				q.Set("authSource", "admin")
				parsed.RawQuery = q.Encode()
			}
			return parsed.String()
		}
	}

	// If no separate credentials, try to build from individual components
	if uri == "" {
		host := os.Getenv("MONGODB_HOST")
		port := os.Getenv("MONGODB_PORT")
		dbName := os.Getenv("MONGODB_DATABASE")
		if host != "" && port != "" && dbName != "" {
			if user != "" && password != "" {
				uri = fmt.Sprintf("mongodb://%s:%s@%s:%s/%s?authSource=admin",
					url.QueryEscape(user), url.QueryEscape(password), host, port, dbName)
			} else {
				uri = fmt.Sprintf("mongodb://%s:%s/%s", host, port, dbName)
			}
		}
	}

	return uri
}

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
			URI:         buildMongoURI(),
			PublicKey:   os.Getenv("MONGO_PUB_KEY"),
			Certificate: os.Getenv("MONGO_CERT"),
			DBName:      os.Getenv("MONGODB_DATABASE"),
		},
	}

	return config, nil
}
