package db

import (
	steam_entity "github.com/replay-api/replay-api/pkg/domain/steam/entities"
	"github.com/resource-ownership/go-mongodb/pkg/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
)

type SteamUserRepository struct {
	mongodb.MongoDBRepository[steam_entity.SteamUser]
}

func NewSteamUserMongoDBRepository(client *mongo.Client, dbName string, entityType steam_entity.SteamUser, collectionName string) *SteamUserRepository {
	repo := mongodb.NewMongoDBRepository[steam_entity.SteamUser](client, dbName, entityType, collectionName, "SteamUser")

	repo.InitQueryableFields(map[string]bool{
		"ID":          true,
		"VHash":       true,
		"SteamID":     true,
		"RealName":    true,
		"PersonaName": true,
	}, map[string]string{
		"ID":          "_id",
		"VHash":       "v_hash",
		"SteamID":     "steam._id",
		"RealName":    "steam.realname",
		"PersonaName": "steam.personaname",
	})

	return &SteamUserRepository{
		MongoDBRepository: *repo,
	}
}
