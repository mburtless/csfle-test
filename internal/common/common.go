package common

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	dbName   = "medicalRecords"
	collName = "patients"
)

func NewMongoClient(uri string, autoEncryptionOpts *options.AutoEncryptionOptions) (*mongo.Client, error) {
	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().
		ApplyURI(uri).
		SetServerAPIOptions(serverAPIOptions)
	if autoEncryptionOpts != nil {
		clientOptions.SetAutoEncryptionOptions(autoEncryptionOpts)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("connect error for regular client: %v", err)
	}
	log.Printf("connected to mongo uri: %s\n", uri)
	return client, nil
}

func GetKMSProviders() (map[string]map[string]interface{}, error) {
	accessKeyID := os.Getenv("MONGODB_ACCESS_KEY_ID")
	if accessKeyID == "" {
		return nil, errors.New("MONGODB_ACCESS_KEY_ID must be set")
	}
	secretAccessKey := os.Getenv("MONGODB_SECRET_ACCESS_KEY")
	if secretAccessKey == "" {
		return nil, errors.New("MONGODB_SECRET_ACCESS_KEY must be set")
	}

	provider := "aws"
	return map[string]map[string]interface{}{
		provider: {
			"accessKeyId":     accessKeyID,
			"secretAccessKey": secretAccessKey,
		},
	}, nil
}

func CreateEncyrptionSchema(dekID string) (map[string]interface{}, error) {
	schemaTemplate := `{
	"bsonType": "object",
	"encryptMetadata": {
		"keyId": [
			{
				"$binary": {
					"base64": "%s",
					"subType": "04"
				}
			}
		]
	},
	"properties": {
		"insurance": {
			"bsonType": "object",
			"properties": {
				"policyNumber": {
					"encrypt": {
						"bsonType": "int",
						"algorithm": "AEAD_AES_256_CBC_HMAC_SHA_512-Deterministic"
					}
				}
			}
		},
		"medicalRecords": {
			"encrypt": {
				"bsonType": "array",
				"algorithm": "AEAD_AES_256_CBC_HMAC_SHA_512-Random"
			}
		},
		"bloodType": {
			"encrypt": {
				"bsonType": "string",
				"algorithm": "AEAD_AES_256_CBC_HMAC_SHA_512-Random"
			}
		},
		"ssn": {
			"encrypt": {
				"bsonType": "int",
				"algorithm": "AEAD_AES_256_CBC_HMAC_SHA_512-Deterministic"
			}
		}
	}
}`
	schema := fmt.Sprintf(schemaTemplate, dekID)
	var schemaDoc bson.Raw
	if err := bson.UnmarshalExtJSON([]byte(schema), true, &schemaDoc); err != nil {
		return nil, fmt.Errorf("UnmarshalExtJSON error: %v", err)
	}
	return map[string]interface{}{
		dbName + "." + collName: schemaDoc,
	}, nil
}
