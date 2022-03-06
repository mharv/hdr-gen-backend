package storage

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"

	"github.com/joho/godotenv"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/streaming"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

var accountName = goDotEnvVariable("AZURE_STORAGE_ACCOUNT_NAME")
var accountKey = goDotEnvVariable("AZURE_STORAGE_PRIMARY_ACCOUNT_KEY")
var containerUrl = goDotEnvVariable("AZURE_STORAGE_CONTAINER_URL")
var containerName = goDotEnvVariable("AZURE_STORAGE_CONTAINER_NAME")
var tmpDirName = goDotEnvVariable("LOCAL_TEMP_DIRECTORY_NAME")

func ConnectBlobStorage() {

	container := connectToStorageContainer(accountName, accountKey)

	downloadFileToLocalDir("test.jpg", tmpDirName, container)

	uploadFileToBlobStore("testUpload.jpg", tmpDirName, container)

}

func connectToStorageContainer(accountName, accountKey string) azblob.ContainerClient {
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

	container := service.NewContainerClient(containerName)

	return container
}

func uploadFileToBlobStore(fileName, directory string, container azblob.ContainerClient) {

	// read file from /tmp
	file, err := os.Open(directory + "/" + fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	fileInfo, _ := file.Stat()
	var size int64 = fileInfo.Size()

	buffer := make([]byte, size)

	// read file content to buffer
	file.Read(buffer)

	fileBytes := bytes.NewReader(buffer)

	// Create a new BlockBlobClient from the ContainerClient
	blockBlob := container.NewBlockBlobClient(fileName)

	// Upload data to the block blob
	_, err = blockBlob.Upload(context.TODO(), streaming.NopCloser(fileBytes), nil)
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Printf("file: %s from directory: %s uploaded to blob\n", fileName, directory)
	}
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
		log.Fatal(err)
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
