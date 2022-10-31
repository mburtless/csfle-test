package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/mburtless/csfle-test/internal/common"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	keyVaultColl      = "__keyVault"
	keyVaultDb        = "encryption"
	keyVaultNamespace = keyVaultDb + "." + keyVaultColl

	dbName   = "medicalRecords"
	collName = "patients"
	testSSN  = 241014209
)

type explainResult struct {
	QueryPlanner queryPlanner `bson:"queryPlanner"`
}

type queryPlanner struct {
	WinningPlan map[string]interface{} `bson:"winningPlan"`
}

func main() {
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatal("MONGODB_URI must be set")
	}
	dekID := os.Getenv("MONGODB_DEK_ID")
	if dekID == "" {
		log.Fatal("MONGODB_DEK_ID must be set")
	}

	schemaMap, err := common.CreateEncyrptionSchema(dekID)
	if err != nil {
		log.Fatal(err)
	}

	kmsProviders, err := common.GetKMSProviders()
	if err != nil {
		log.Fatal(err)
	}
	autoEncryptionOpts := options.AutoEncryption().
		SetKmsProviders(kmsProviders).
		SetKeyVaultNamespace(keyVaultNamespace).
		SetSchemaMap(schemaMap)

	regularClient, err := common.NewMongoClient(uri, nil)
	if err != nil {
		log.Fatal(err)
	}
	secureClient, err := common.NewMongoClient(uri, autoEncryptionOpts)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = regularClient.Disconnect(context.TODO())
		_ = secureClient.Disconnect(context.TODO())
	}()

	fmt.Println("Finding a document with regular (non-encrypted) client.")
	var resultRegular bson.M
	err = regularClient.Database(dbName).Collection(collName).FindOne(context.TODO(), bson.D{{"name", "Jon Doe"}}).Decode(&resultRegular)
	if err != nil {
		log.Fatal(err)
	}
	outputRegular, err := json.MarshalIndent(resultRegular, "", "    ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", outputRegular)
	fmt.Println("Finding a document with encrypted client, searching on an encrypted field")
	var resultSecure bson.M
	err = secureClient.Database(dbName).Collection(collName).FindOne(context.TODO(), bson.D{{"ssn", testSSN}}).Decode(&resultSecure)
	if err != nil {
		log.Fatal(err)
	}
	outputSecure, err := json.MarshalIndent(resultSecure, "", "    ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", outputSecure)

	fmt.Println("Show query plan for find on indexed encrypted field")
	filter := bson.D{{"ssn", testSSN}}
	findCmd := bson.D{{"find", collName}, {"filter", filter}}
	command := bson.D{{"explain", findCmd}}
	var result explainResult
	err = secureClient.Database(dbName).RunCommand(context.TODO(), command).Decode(&result)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Query plan: %s\n", result)
}
