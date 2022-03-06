package main

import (
	"log"

	"hdr-gen-backend/database"
	"hdr-gen-backend/router"
	"hdr-gen-backend/storage"
)

func main() {

	database.ConnectDatabase()

	storage.ConnectBlobStorage()

	routerInstance := router.NewRouter()

	log.Fatal(routerInstance.Run(":8080"))

	// r.SetTrustedProxies([]string{"127.0.0.1"})

}
