package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"

	"github.com/mburtless/csfle-test/internal/common"
	"github.com/moby/moby/pkg/namesgenerator"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	keyVaultColl      = "__keyVault"
	keyVaultDb        = "encryption"
	keyVaultNamespace = keyVaultDb + "." + keyVaultColl

	dbName   = "medicalRecords"
	collName = "patients"
	provider = "aws"
)

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
	//SetExtraOptions(extraOptions)
	client, err := common.NewMongoClient(uri, autoEncryptionOpts)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = client.Disconnect(context.TODO())
	}()

	err = insertEncryptedDoc(client)
	if err != nil {
		log.Fatal(err)
	}
}

func insertEncryptedDoc(client *mongo.Client) error {
	// fields will be automatically encrypted on insert
	// according to encryption schema
	var testPatients []interface{}
	for i := 0; i < 20; i++ {
		testPatient := map[string]interface{}{
			"name":      namesgenerator.GetRandomName(0),
			"ssn":       rand.Intn(999999999-100000000) + 100000000,
			"bloodType": "AB+",
			"medicalRecords": []map[string]interface{}{{
				"weight":        180,
				"bloodPressure": "120/80",
			}},
			"insurance": map[string]interface{}{
				"provider":     "MaestCare",
				"policyNumber": 123142,
			},
		}
		testPatients = append(testPatients, testPatient)
	}
	res, err := client.Database(dbName).Collection(collName).InsertMany(context.TODO(), testPatients)
	if err != nil {
		return fmt.Errorf("InsertOne error: %v", err)
	}
	fmt.Printf("inserted patients with IDs %v\n", res.InsertedIDs)

	return nil
}
