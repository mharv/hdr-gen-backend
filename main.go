package main

import (
	"log"

	"hdr-gen-backend/database"
	"hdr-gen-backend/router"
)

func main() {

	database.ConnectDatabase()

	routerInstance := router.NewRouter()

	log.Fatal(routerInstance.Run(":8080"))

	// r.SetTrustedProxies([]string{"127.0.0.1"})

}
