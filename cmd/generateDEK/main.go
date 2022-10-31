package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/mburtless/csfle-test/internal/common"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	cmkRegion = "us-east-1"

	keyVaultColl      = "__keyVault"
	keyVaultDb        = "encryption"
	keyVaultNamespace = keyVaultDb + "." + keyVaultColl
	provider          = "aws"
)

func main() {
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatal("MONGODB_URI must be set")
	}
	cmkARN := os.Getenv("MONGODB_CMK_ARN")
	if cmkARN == "" {
		log.Fatal("MONGODB_CMK_ARN must be set")
	}

	//keyVaultClient, err := newKeyVaultClient(uri)
	keyVaultClient, err := common.NewMongoClient(uri, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = keyVaultClient.Disconnect(context.TODO())
	}()

	if err := createDatabases(keyVaultClient); err != nil {
		log.Fatal(err)
	}

	dek, err := makeDataKey(keyVaultClient)
	if err != nil {
		log.Fatal(err)
	}
	dekData := base64.StdEncoding.EncodeToString(dek.Data)
	fmt.Printf("DataKeyId [base64]: %s\nexport MONGODB_DEK_ID=%s\n", dekData, dekData)
}

// drops keyvault and medicalRecords dbs from previous demos and creates new ones
func createDatabases(keyVaultClient *mongo.Client) error {
	// Drop the Key Vault Collection in case you created this collection
	// in a previous run of this application.
	log.Println("dropping previous keyvault db")
	if err := keyVaultClient.Database(keyVaultDb).Collection(keyVaultColl).Drop(context.TODO()); err != nil {
		return fmt.Errorf("Collection.Drop error: %v", err)
	}

	// Drop the database storing your encrypted fields as all
	// the DEKs encrypting those fields were deleted in the preceding line.
	log.Println("dropping previous medicalRecords db")
	if err := keyVaultClient.Database("medicalRecords").Collection("patients").Drop(context.TODO()); err != nil {
		return fmt.Errorf("Collection.Drop error: %v", err)
	}

	log.Println("creating new keyvault db")
	// unique index on keyAltNames for keyVault collection
	keyVaultIndex := mongo.IndexModel{
		Keys: bson.D{{"keyAltNames", 1}},
		Options: options.Index().
			SetUnique(true).
			SetPartialFilterExpression(bson.D{
				{"keyAltNames", bson.D{
					{"$exists", true},
				}},
			}),
	}
	_, err := keyVaultClient.Database(keyVaultDb).Collection(keyVaultColl).Indexes().CreateOne(context.TODO(), keyVaultIndex)
	if err != nil {
		return errors.New(err.Error())
	}
	log.Println("keyvault db created")

	log.Println("creating new medicalRecords db")
	// example index on an encrypted field
	medicalRecordsIndex := mongo.IndexModel{
		Keys:    bson.D{{"ssn", 1}},
		Options: nil,
	}
	_, err = keyVaultClient.Database("medicalRecords").Collection("patients").Indexes().CreateOne(context.TODO(), medicalRecordsIndex)
	if err != nil {
		return errors.New(err.Error())
	}
	log.Println("medicalRecords db created")

	return nil
}

func makeDataKey(keyVaultClient *mongo.Client) (*primitive.Binary, error) {
	log.Println("creating DEK")

	kmsProviders, err := common.GetKMSProviders()
	if err != nil {
		return nil, err
	}

	masterKey := map[string]interface{}{
		"key":    cmkARN,
		"region": cmkRegion,
	}
	clientEncryptionOpts := options.ClientEncryption().SetKeyVaultNamespace(keyVaultNamespace).
		SetKmsProviders(kmsProviders)
	clientEnc, err := mongo.NewClientEncryption(keyVaultClient, clientEncryptionOpts)
	if err != nil {
		return nil, fmt.Errorf("NewClientEncryption error %v", err)
	}
	defer func() {
		_ = clientEnc.Close(context.TODO())
	}()
	dataKeyOpts := options.DataKey().
		SetMasterKey(masterKey)
	dataKeyID, err := clientEnc.CreateDataKey(context.TODO(), provider, dataKeyOpts)
	if err != nil {
		return nil, fmt.Errorf("create data key error %v", err)
	}
	return &dataKeyID, nil
}
