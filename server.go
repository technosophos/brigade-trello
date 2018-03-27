package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/Azure/brigade/pkg/brigade"
	"github.com/Azure/brigade/pkg/storage"
	"github.com/Azure/brigade/pkg/storage/kube"
	"github.com/Azure/brigade/pkg/webhook"

	"github.com/gin-gonic/gin"
)

// EnvFetchScript names the environment veriable for script fetching.
//
// If this environment variable is set to a truthy value, then this gateway
// will try to use the GitHub API to fetch brigade script rather than leaving
// it to the controller to load the script.
const EnvFetchScript = "BRIGADE_FETCH_SCRIPT"

var store storage.Store

func main() {

	client, err := kube.GetClient("", os.Getenv("KUBECONFIG"))
	if err != nil {
		panic(err)
	}
	store = kube.New(client, "default") // FIXME: accept alt namespaces

	//router := gin.Default()
	router := gin.New()
	router.Use(gin.Recovery())
	router.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "ok"}) })

	// Trello uses GET/HEAD to test connection
	trello := router.Group("/trello")
	trello.Use(gin.Logger())
	trello.GET("/:project", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "ok"}) })
	trello.HEAD("/:project", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "ok"}) })
	trello.POST("/:project", trelloFn)

	// Generic is just a generic webhook handler
	generic := router.Group("/generic")
	generic.Use(gin.Logger())
	generic.GET("/:project", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "ok"}) })
	generic.POST("/:project", genericFn)

	router.Run()
}

func trelloFn(c *gin.Context) {
	pid := c.Param("project")

	// INCOMPLETE: This is the first step in validating the request originated
	// from Trello. Finish. https://developers.trello.com/page/webhooks
	sig := c.Request.Header.Get("x-trello-webhook")
	if sig == "" {
		log.Println("No X-Trello-Webhook header present. Skipping")
		c.JSON(http.StatusBadRequest, gin.H{"status": "Malformed headers"})
		return
	}
	// TODO: validate that the body matches the hash in the sig header.

	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "Malformed body"})
		return
	}
	c.Request.Body.Close()

	// Load project
	proj, err := store.GetProject(pid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "Resource Not Found"})
		return
	}

	// Create the build
	build := &brigade.Build{
		ProjectID: pid,
		Type:      "trello",
		Provider:  "trello",
		Revision: &brigade.Revision{
			Ref: "master",
		},
		Payload: body,
	}

	if fetch, ok := os.LookupEnv(EnvFetchScript); ok && fetch == "1" {
		// Get the brigade.js
		// Right now, we skip this and let the github project handle it.
		script, err := webhook.GetFileContents(proj, "master", "brigade.js")
		if err != nil {
			log.Printf("Error getting file: %s", err)
			c.JSON(http.StatusNotFound, gin.H{"status": "Script Not Found"})
			return
		}
		build.Script = script
	}

	if err := store.CreateBuild(build); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "Failed to invoke hook"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OK"})
}

func genericFn(c *gin.Context) {
	log.Printf("generic webhook from user-agent %q", c.Request.UserAgent())
	pid := c.Param("project")

	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "Malformed body"})
		return
	}
	c.Request.Body.Close()

	// Load project
	proj, err := store.GetProject(pid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "Resource Not Found"})
		return
	}

	// Get the brigade.js
	// Right now, we skip this and let the github project handle it.
	script, err := webhook.GetFileContents(proj, "master", "brigade.js")
	if err != nil {
		log.Printf("Error getting file: %s", err)
		c.JSON(http.StatusNotFound, gin.H{"status": "Script Not Found"})
		return
	}

	// Create the build
	build := &brigade.Build{
		ProjectID: pid,
		Type:      "webhook",
		Provider:  c.Request.UserAgent(),
		Revision: &brigade.Revision{
			Ref: "master",
		},
		Payload: body,
		Script:  script,
	}
	if err := store.CreateBuild(build); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "Failed to invoke hook"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OK"})
}
