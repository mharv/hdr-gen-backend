package router

import (
	"net/http"

	"hdr-gen-backend/handlers"

	"github.com/gin-gonic/gin"
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

// NewRouter returns a new router.
func NewRouter() *gin.Engine {
	router := gin.Default()
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
		"GetProjects",
		http.MethodGet,
		"/projects",
		handlers.GetProjects,
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
}
