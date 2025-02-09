package iam_entities

import "github.com/google/uuid"

type Jwk struct {
	Kid        uuid.UUID `json:"kid" bson:"_id"`
	Kty        string    `json:"kty" bson:"kty"`
	E          string    `json:"e" bson:"e"`
	N          string    `json:"n" bson:"n"`
	Use        string    `json:"use" bson:"use"`
	Alg        string    `json:"alg" bson:"alg"`
	PrivateKey string    `json:"-" bson:"private_key"`
}

func (e Jwk) GetID() uuid.UUID {
	return e.Kid
}
