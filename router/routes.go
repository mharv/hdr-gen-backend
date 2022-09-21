package router

import (
	"log"
	"net/http"
	"os"

	"hdr-gen-backend/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/joho/godotenv"
)

// Route is the information for every URI.
type Route struct {
	// Name is the name of this Route.
	Name string
	// Method is the string for the HTTP method. ex) GET, POST etc..
	Method string
	// Pattern is the pattern of the URI.
	Pattern string
	// HandlerFunc is the handler function of this route.
	HandlerFunc gin.HandlerFunc
}

type Routes []Route

var deployedUrl = goDotEnvVariable("FRONTEND_URL")

// NewRouter returns a new router.
func NewRouter() *gin.Engine {
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:3000", deployedUrl},
		AllowMethods: []string{"POST", "GET", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Access-Control-Allow-Headers", "content-type"},
		// ExposeHeaders: []string{"Content-Length"},
	}))
	for _, route := range routes {
		switch route.Method {
		case http.MethodGet:
			router.GET(route.Pattern, route.HandlerFunc)
		case http.MethodPost:
			router.POST(route.Pattern, route.HandlerFunc)
		case http.MethodPut:
			router.PUT(route.Pattern, route.HandlerFunc)
		case http.MethodPatch:
			router.PATCH(route.Pattern, route.HandlerFunc)
		case http.MethodDelete:
			router.DELETE(route.Pattern, route.HandlerFunc)
		}
	}

	return router
}

var routes = Routes{
	{
		"GetProjectByNumber",
		http.MethodGet,
		"/projectByNumber/:projectNumber",
		handlers.GetProjectByNumber,
	},
	{
		"GetProjects",
		http.MethodGet,
		"/projects",
		handlers.GetProjects,
	},
	{
		"PostLuminanceAverages",
		http.MethodPost,
		"/luminanceAverages/:projectId/:imageName",
		handlers.PostLuminanceAverages,
	},
	{
		"PostProject",
		http.MethodPost,
		"/projects",
		handlers.PostProject,
	},

	{
		"GetImages",
		http.MethodGet,
		"/images",
		handlers.GetImages,
	},

	{
		"GetImageByName",
		http.MethodGet,
		"/imageByName/:imageName",
		handlers.GetImageByName,
	},

	{
		"GetImagesProjectId",
		http.MethodGet,
		"/imagesByProject/:projectId",
		handlers.GetImagesProjectId,
	},

	{
		"DownloadImagesProjectId",
		http.MethodGet,
		"/dowloadImagesByProject/:projectId",
		handlers.DownloadImagesProjectId,
	},

	{
		"UploadImagesToServer",
		http.MethodPost,
		"/uploadImages/:projectId/:imageName",
		handlers.UploadImagesToServer,
	},
	{
		"UpExposeImage",
		http.MethodGet,
		"/upExposeImage/:projectId/:imageName/:exposureFactor",
		handlers.UpExposeImage,
	},
	{
		"DownExposeImage",
		http.MethodGet,
		"/downExposeImage/:projectId/:imageName/:exposureFactor",
		handlers.DownExposeImage,
	},
	{
		"LuminanceLevels",
		http.MethodGet,
		"/luminanceLevels/:projectId/:imageName",
		handlers.LuminanceLevels,
	},

	{
		"LuminanceMatrix",
		http.MethodGet,
		"/luminanceMatrix/:projectId/:imageName",
		handlers.LuminanceMatrix,
	},
	{
		"ScaleImage",
		http.MethodGet,
		"/scaleImage/:projectId/:imageName/:current/:target",
		handlers.ScaleImage,
	},
	{
		"FalseColour",
		http.MethodGet,
		"/falseColour/:projectId/:imageName",
		handlers.FalseColour,
	},
	{
		"DeleteProjectByNumber",
		http.MethodGet,
		"/deleteProjectByNumber/:projectNumber",
		handlers.DeleteProjectByNumber,
	},
	//test routes
	{
		"Rtrace",
		http.MethodGet,
		"/rtrace",
		handlers.Rtrace,
	},
	{
		"Sleep",
		http.MethodGet,
		"/sleep",
		handlers.Sleep,
	},
	{
		"ApplogTest",
		http.MethodGet,
		"/applogtest",
        handlers.TestLog,
	},
}

func goDotEnvVariable(key string) string {

	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv(key)
}
