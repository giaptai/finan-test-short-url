package main

import (
	"github.com/giaptai/finan-test-short-url/bll"
	"github.com/giaptai/finan-test-short-url/dao"
	"github.com/giaptai/finan-test-short-url/dao/database"
	"github.com/giaptai/finan-test-short-url/handler"
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	db, err := database.Connection()
	if err != nil {
		log.Fatal("Cannot connect to database:", err)
	}
	defer db.Close()

	repo := dao.NewURLRepository(db)
	service := bll.NewURLService(repo)
	handler := handler.NewURLHandler(service, "http://localhost:8088")

	// create router
	r := gin.Default()
	r.POST("/api/urls", handler.CreateShortURL)
	r.GET("/api/urls/:shortCode", handler.GetURLInfo)
	r.GET("/api/urls", handler.ListURLs)
	r.GET("/:shortCode", handler.RedirectToOriginal)
	r.Run(":8088")
}
