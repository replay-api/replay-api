package db

import (
	"reflect"

	steam_entity "github.com/replay-api/replay-api/pkg/domain/steam/entities"
	"go.mongodb.org/mongo-driver/mongo"
)

type SteamUserRepository struct {
	MongoDBRepository[steam_entity.SteamUser]
}

func NewSteamUserMongoDBRepository(client *mongo.Client, dbName string, entityType steam_entity.SteamUser, collectionName string) *SteamUserRepository {
	repo := MongoDBRepository[steam_entity.SteamUser]{
		mongoClient:       client,
		dbName:            dbName,
		mappingCache:      make(map[string]CacheItem),
		entityModel:       reflect.TypeOf(entityType),
		BsonFieldMappings: make(map[string]string),
		collectionName:    collectionName,
		entityName:        reflect.TypeOf(entityType).Name(),
		QueryableFields:   make(map[string]bool),
	}

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
		repo,
	}
}
