package main

import (
	"blog-api/internal/config"
	"blog-api/internal/handlers"
	"blog-api/internal/repository"
	"blog-api/internal/services"
	"database/sql"
	"log"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Setup database
	db, err := sql.Open("postgres", cfg.DatabaseURL())
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test database connection with retry
	for i := 0; i < 10; i++ {
		if err = db.Ping(); err == nil {
			break
		}
		log.Printf("Failed to ping database (attempt %d): %v", i+1, err)
		time.Sleep(time.Second * 2)
	}
	if err != nil {
		log.Fatal("Could not establish database connection:", err)
	}

	// Setup Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr(),
	})

	// Setup Elasticsearch
	esCfg := elasticsearch.Config{
		Addresses: []string{cfg.ElasticsearchURL},
	}
	esClient, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		log.Fatal("Failed to create Elasticsearch client:", err)
	}

	// Setup repositories and services
	postRepo := repository.NewPostRepository(db)
	activityLogRepo := repository.NewActivityLogRepository(db)
	cacheService := services.NewCacheService(rdb)
	searchService := services.NewSearchService(esClient)
	
	postService := services.NewPostService(postRepo, activityLogRepo, cacheService, searchService)

	// Setup handlers
	postHandler := handlers.NewPostHandler(postService)

	// Setup router
	r := gin.Default()

	// Routes
	api := r.Group("/api/v1")
	{
		api.POST("/posts", postHandler.CreatePost)
		api.GET("/posts/:id", postHandler.GetPost)
		api.PUT("/posts/:id", postHandler.UpdatePost)
		api.GET("/posts/search-by-tag", postHandler.SearchByTag)
		api.GET("/posts/search", postHandler.SearchPosts)
	}

	// Start server
	log.Printf("Server starting on port %s", cfg.Port)
	log.Fatal(r.Run(":" + cfg.Port))
}