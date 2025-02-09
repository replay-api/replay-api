package db

import (
	"reflect"

	iam_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/entities"
	"go.mongodb.org/mongo-driver/mongo"
)

type JwtRepository struct {
	MongoDBRepository[iam_entities.Jwk]
}

func NewJwtRepository(client *mongo.Client, dbName string, entityType *iam_entities.Jwk, collectionName string) *JwtRepository {
	repo := MongoDBRepository[iam_entities.Jwk]{
		mongoClient:       client,
		dbName:            dbName,
		mappingCache:      make(map[string]CacheItem),
		entityModel:       reflect.TypeOf(entityType),
		bsonFieldMappings: make(map[string]string),
		collectionName:    collectionName,
	}

	repo.InitQueryableFields(map[string]bool{
		"Kid":        true,
		"Kty":        false,
		"E":          false,
		"N":          false,
		"Use":        false,
		"Alg":        false,
		"PrivateKey": false,
	}, map[string]string{
		"Kid":        "_id",
		"Kty":        "kty",
		"E":          "e",
		"N":          "n",
		"Use":        "use",
		"Alg":        "alg",
		"PrivateKey": "private_key",
	})

	return &JwtRepository{
		repo,
	}
}
