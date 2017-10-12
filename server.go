package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/deis/brigade/pkg/brigade"
	"github.com/deis/brigade/pkg/storage"
	"github.com/deis/brigade/pkg/storage/kube"
	"github.com/deis/brigade/pkg/webhook"

	"github.com/gin-gonic/gin"
)

var store storage.Store

func main() {

	client, err := kube.GetClient("", os.Getenv("KUBECONFIG"))
	if err != nil {
		panic(err)
	}
	store = kube.New(client, "default") // FIXME: accept alt namespaces

	router := gin.Default()

	router.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "ok"}) })

	// Trello uses GET/HEAD to test connection
	router.GET("/trello/:project", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "ok"}) })
	router.HEAD("/trello/:project", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "ok"}) })
	router.POST("/trello/:project", trello)

	router.Run()
}

func trello(c *gin.Context) {
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
		Type:      "trello",
		Provider:  "trello",
		Commit:    "master",
		Payload:   body,
		Script:    script,
	}
	if err := store.CreateBuild(build); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "Failed to invoke hook"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OK"})
}
