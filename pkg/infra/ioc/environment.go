package ioc

import (
	"os"

	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
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
			Issuer: os.Getenv("ISSUER"),
		},
		MongoDB: common.MongoDBConfig{
			URI:         os.Getenv("MONGO_URI"),
			PublicKey:   os.Getenv("MONGO_PUB_KEY"),
			Certificate: os.Getenv("MONGO_CERT"),
			DBName:      os.Getenv("MONGO_DB_NAME"),
		},
		Api: common.ApiConfig{
			Port: os.Getenv("PORT"),
		},
	}

	return config, nil
}
