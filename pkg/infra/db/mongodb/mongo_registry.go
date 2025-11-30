package db

import (
	"fmt"
	"reflect"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
)

var (
	tUUID       = reflect.TypeOf(uuid.UUID{})
	uuidSubtype = byte(0x04)

	MongoRegistry = newUUIDRegistry()
)

func newUUIDRegistry() *bsoncodec.Registry {
	reg := bson.NewRegistry()
	reg.RegisterTypeEncoder(tUUID, bsoncodec.ValueEncoderFunc(uuidEncodeValue))
	reg.RegisterTypeDecoder(tUUID, bsoncodec.ValueDecoderFunc(uuidDecodeValue))
	return reg
}

func uuidEncodeValue(ec bsoncodec.EncodeContext, vw bsonrw.ValueWriter, val reflect.Value) error {
	if !val.IsValid() || val.Type() != tUUID {
		return bsoncodec.ValueEncoderError{Name: "uuidEncodeValue", Types: []reflect.Type{tUUID}, Received: val}
	}
	b := val.Interface().(uuid.UUID)
	return vw.WriteBinaryWithSubtype(b[:], uuidSubtype)
}

func uuidDecodeValue(dc bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value) error {
	if !val.CanSet() || val.Type() != tUUID {
		return bsoncodec.ValueDecoderError{Name: "uuidDecodeValue", Types: []reflect.Type{tUUID}, Received: val}
	}

	var data []byte
	var subtype byte
	var err error
	switch vrType := vr.Type(); vrType {
	case bson.TypeBinary:
		data, subtype, err = vr.ReadBinary()
		if subtype != uuidSubtype {
			return fmt.Errorf("unsupported binary subtype %v for UUID", subtype)
		}
	case bson.TypeNull:
		err = vr.ReadNull()
	case bson.TypeUndefined:
		err = vr.ReadUndefined()
	default:
		return fmt.Errorf("cannot decode %v into a UUID", vrType)
	}

	if err != nil {
		return err
	}
	uuid2, err := uuid.FromBytes(data)
	if err != nil {
		return err
	}
	val.Set(reflect.ValueOf(uuid2))
	return nil
}
