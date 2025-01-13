package db

import (
	"reflect"

	steam_entity "github.com/psavelis/team-pro/replay-api/pkg/domain/steam/entities"
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
		bsonFieldMappings: make(map[string]string),
		collectionName:    collectionName,
		entityName:        reflect.TypeOf(entityType).Name(),
		queryableFields:   make(map[string]bool),
	}

	repo.InitQueryableFields(map[string]bool{
		"ID":                true,
		"VHash":             true,
		"Steam.*":           true,
		"Steam.RealName":    true,
		"Steam.PersonaName": true,
	}, map[string]string{
		"ID":                "_id",
		"VHash":             "v_hash",
		"Steam":             "steam",
		"Steam.ID":          "steam._id",
		"Steam.RealName":    "steam.realname",
		"Steam.PersonaName": "steam.personaname",
	})

	return &SteamUserRepository{
		repo,
	}
}
