package storage

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/joho/godotenv"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/streaming"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

// https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/storage/azblob#BlockBlobClient.Upload

var accountName = goDotEnvVariable("AZURE_STORAGE_ACCOUNT_NAME")
var accountKey = goDotEnvVariable("AZURE_STORAGE_PRIMARY_ACCOUNT_KEY")
var containerUrl = goDotEnvVariable("AZURE_STORAGE_CONTAINER_URL")
var containerName = goDotEnvVariable("AZURE_STORAGE_CONTAINER_NAME")
var tmpDirName = goDotEnvVariable("LOCAL_TEMP_DIRECTORY_NAME")

func ConnectBlobStorage() {

	// testing blob store

	// container := ConnectToStorageContainer(accountName, accountKey)
	// file, err := os.Open("./scripts/test.txt")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer file.Close()

	// fileInfo, _ := file.Stat()
	// var size int64 = fileInfo.Size()

	// buffer := make([]byte, size)

	// file.Read(buffer)

	// fileBytes := bytes.NewReader(buffer)

	// // Create a new BlockBlobClient from the ContainerClient
	// blockBlob := container.NewBlockBlobClient("test.txt")

	// // Upload data to the block blob
	// _, err = blockBlob.Upload(context.TODO(), streaming.NopCloser(fileBytes), nil)
	// if err != nil {
	// 	log.Fatal(err)
	// 	fmt.Sprintf("%s was not saved to the blob store", "test.txt")
	// } else {
	// 	fmt.Printf("file: %s  uploaded to blob \n", "test.txt")
	// }

}

func ConnectToStorageContainer(accountName, accountKey string) azblob.ContainerClient {
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

func UploadFileToBlobStore(fileName, directory string, uuidRequired bool) string {
	container := ConnectToStorageContainer(accountName, accountKey)

	// read file from /tmp
	file, err := os.Open(directory + fileName)
	fmt.Println(directory + fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	fileInfo, _ := file.Stat()
	var size int64 = fileInfo.Size()

	buffer := make([]byte, size)

	extension := filepath.Ext(fileName)
	fileNameOnly := strings.TrimSuffix(fileName, extension)
	var blobFileName string

	if uuidRequired {
		blobFileName = fileNameOnly + "-" + uuid.New().String() + extension
	} else {
		blobFileName = fileNameOnly + extension
	}

	// read file content to buffer
	file.Read(buffer)

	fileBytes := bytes.NewReader(buffer)

	// Create a new BlockBlobClient from the ContainerClient
	blockBlob := container.NewBlockBlobClient(blobFileName)

	// Upload data to the block blob
	_, err = blockBlob.Upload(context.TODO(), streaming.NopCloser(fileBytes), nil)
	if err != nil {
		log.Fatal(err)
		return fmt.Sprintf("%s was not saved to the blob store", blobFileName)
	} else {
		fmt.Printf("file: %s from directory: %s uploaded to blob as: %s \n", fileName, directory, blobFileName)
		return blobFileName
	}
}

func DownloadFileToLocalDir(fileName, directory string) {
	container := ConnectToStorageContainer(accountName, accountKey)

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

	err = ioutil.WriteFile(directory+fileName, downloadedData.Bytes(), fs.FileMode(permissions))
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
