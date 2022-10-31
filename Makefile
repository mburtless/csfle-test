build-generateDEK:
	go build -tags cse -o generateDEK cmd/generateDEK/main.go

build-insertEncryptedDoc:
	go build -tags cse -o insertEncryptedDoc cmd/insertEncryptedDoc/main.go

build-retrieveEncryptedDoc:
	go build -tags cse -o retrieveEncryptedDoc cmd/retrieveEncryptedDoc/main.go

build: build-generateDEK
build: build-insertEncryptedDoc
build: build-retrieveEncryptedDoc
