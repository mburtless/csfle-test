# csfle-test
Testing Mongo CSFLE with AWS KMS, based on the instructions in the [Mongo docs](https://www.mongodb.com/docs/manual/core/csfle/tutorials/aws/aws-automatic/)

# Pre-reqs
* A Mongo Enterprise instance or [Atlas](https://www.mongodb.com/try) Cluster. Atlas is easy to setup, hosted, and free.
* `mongocryptd` binary in `PATH`.  This can be obtained from the [Mongo Enterprise Server](https://www.mongodb.com/try/download/enterprise) tgz.
* Install `libmongocrypt`: `brew install mongodb/brew/libmongocrypt`

# Set Up the KMS
1. Follow the steps in [Set Up the KMS](https://www.mongodb.com/docs/manual/core/csfle/tutorials/aws/aws-automatic/#set-up-the-kms) section of the mongo docs.
2. `export MONGODB_CMK_ARN=<ARN of CMK created above>`

# Create an AWS IAM User
1. Follow the steps in [Create an AWS IAM User](https://www.mongodb.com/docs/manual/core/csfle/tutorials/aws/aws-automatic/#create-an-aws-iam-user) section of the Mongo docs.
2. `export MONGODB_ACCESS_KEY_ID=<access key ID of IAM user created above>`
3. `export MONGODB_SECRET_ACCESS_KEY=<secret access key of IAM user created above>`

# Run demo
1. `make build`
2. Create demo DBs a new Data Enryption Key: `./generateDEK`
3. `export MONGODB_DEK_ID=<DEK ID created above>`
4. Insert data encrypted using DEK: `./insertEncryptedDoc`
5. Retrieve encrypted data: `./retrieveEncryptedDoc`
