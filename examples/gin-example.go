package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	sigsci "github.com/signalsciences/sigsci-module-golang"
)

func hellofunc(c *gin.Context) {
	c.String(http.StatusOK, "hello handler")
}

func main() {
	r := gin.New()
	_, err := sigsci.Middleware(r,
		//sigsci.Socket("unix", "/tmp/sigsci.sock"),
		sigsci.Debug(true),
		sigsci.MaxContentLength(20),
	)
	if err != nil {
		log.Fatal(err)
	}
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	r.GET("/hello", hellofunc)
	s := &http.Server{
		Handler: r,
		Addr:    "127.0.0.1:8080",
	}
	log.Printf("Server URL: http://%s/", s.Addr)
	log.Fatal(s.ListenAndServe())
}
