package main

import (
	"context"
	"demo/db"
	"errors"
	"fmt"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {	
		t := time.Now()

		// before request
		c.Next()

		// after request
		latency := time.Since(t)

		fmt.Printf("Latency: %s\n", latency)

		// access the router information
		route := c.Request.URL.Path
		fmt.Printf("Route: %s\n", route)

		// access the status we are sending
		status := c.Writer.Status()
		fmt.Printf("Status: %v\n", status)
	}
}

// Middleware that checks for API Key in Request Header. If found, it will continue to next() function. If not found, it will return an error.

var allowedKeys = []string{"ELITE"}

func contains(slice []string, item string) bool {
    for _, v := range slice {
        if v == item {
            return true
        }
    }
    return false
}

func ApiKeyAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("SPITFIRE-API-KEY")

		if key == "" {
			c.JSON(401, gin.H{
				"message": "Unauthorized",
			})
			c.Abort()
			return
		}

		if !contains(allowedKeys, key) {
			c.JSON(401, gin.H{
				"message": "Unauthorized",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

type PostInput struct {
	Title       string `json:"title"`
	Published   bool   `json:"published"`
	Description string `json:"description"`
}

func main() {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders: []string{"*"},
		MaxAge: 12 * time.Hour,
		AllowCredentials: true,
	}))

	client := db.NewClient()
	if err := client.Prisma.Connect(); err != nil {
		panic(err)
	}

	defer func() {
		if err := client.Prisma.Disconnect(); err != nil {
			panic(err)
		}
	}()

	r.Use(Logger())
	r.Use(ApiKeyAuthMiddleware())

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello World!",
		})
	})

	// Ping Route
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Server is running!",
		})
	})

	// Get Posts
	r.GET("/posts", func(c *gin.Context) {
		ctx := context.Background()
		posts, err := client.Post.FindMany().Exec(ctx)
		if err != nil {
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
		}
		c.JSON(200, posts)
	})

	// Create Post
	r.POST("/posts", func(c *gin.Context) {
		var input PostInput
		if err := c.BindJSON(&input); err != nil {
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
		}

		title := input.Title
		published := input.Published 
		description := input.Description

		if title == "" {
			err := errors.New("title cannot be empty")
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx := context.Background()

		post, err := client.Post.CreateOne(
			db.Post.Title.Set(title),
			db.Post.Published.Set(published),
			db.Post.Desc.Set(description),
		).Exec(ctx)

		if err != nil {
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
		}

		c.JSON(200, post)
	})

	// Update Post
	r.PUT("/posts/:id", func(c *gin.Context) {
		var input PostInput
		if err := c.BindJSON(&input); err != nil {
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
			return
		}

		title := input.Title
		published := input.Published 
		description := input.Description

		if title == "" {
			err := errors.New("title cannot be empty")
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx := context.Background()
		postID := c.Param("id")

		post, err := client.Post.FindMany(
			db.Post.ID.Equals(postID),
		).Update(
			db.Post.Title.Set(title),
			db.Post.Published.Set(published),
			db.Post.Desc.Set(description),
		).Exec(ctx)

		if err != nil {
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(200, post)
	})

	// Delete Post
	r.DELETE("/posts/:id", func(c *gin.Context) {
		ctx := context.Background()
		postID := c.Param("id")

		post, err := client.Post.FindUnique(
			db.Post.ID.Equals(postID),
		).Delete().Exec(ctx)

		if err != nil {
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
		}

		c.JSON(200, post)
	})

	// 
	r.Run(":1234")
}
