package storage

import (
	"bytes"
	"context"
	"io/fs"
	"io/ioutil"

	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

var accountName = goDotEnvVariable("AZURE_STORAGE_ACCOUNT_NAME")
var accountKey = goDotEnvVariable("AZURE_STORAGE_PRIMARY_ACCOUNT_KEY")

var containerUrl = goDotEnvVariable("AZURE_STORAGE_CONTAINER_URL")
var containerName = goDotEnvVariable("AZURE_STORAGE_CONTAINER_NAME")

func ConnectBlobStorage() {

	service := connectToStorage(accountName, accountKey)

	container := service.NewContainerClient(containerName)

	downloadFileToLocalDir("test.jpg", "/tmp", container)

}

func connectToStorage(accountName, accountKey string) azblob.ServiceClient {
	cred, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("blob connection successful...")
	}

	// The service URL for blob endpoints is usually in the form: http(s)://<account>.blob.core.windows.net/
	service, err := azblob.NewServiceClientWithSharedKey(fmt.Sprintf("https://%s.blob.core.windows.net/", accountName), cred, nil)
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println(service)
	}
	return service
}

func downloadFileToLocalDir(fileName, directory string, container azblob.ContainerClient) {

	// Create a new BlockBlobClient from the ContainerClient
	blockBlob := container.NewBlockBlobClient(fileName)

	get, err := blockBlob.Download(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}

	// Use the bytes.Buffer object to read the downloaded data.
	downloadedData := &bytes.Buffer{}
	reader := get.Body(nil) // RetryReaderOptions has a lot of in-depth tuning abilities, but for the sake of simplicity, we'll omit those here.
	_, err = downloadedData.ReadFrom(reader)
	if err != nil {
		log.Fatal(err)
	}

	err = reader.Close()
	if err != nil {
		log.Fatal(err)
	}

	permissions := 0644

	err = ioutil.WriteFile(directory+"/"+fileName, downloadedData.Bytes(), fs.FileMode(permissions))
	if err != nil {
		// handle error
	} else {
		fmt.Printf("file: %s downloaded to directory: %s \n", fileName, directory)
	}
}

func goDotEnvVariable(key string) string {

	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv(key)
}
