package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	response "example.com/mod/core"
	"example.com/mod/utils"
)

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

var (
	clientOptions *options.ClientOptions
	database      *mongo.Database
	env           = flag.String("env", "dev", "the environment to use (dev, test, prod)")
)

func main() {

	flag.Parse()

	r := gin.New()

	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(*env, "env---")

	database := client.Database("blog")

	r.Use(func(c *gin.Context) {
		c.Set("db", database)
		c.Next()
	})

	r.GET("/", func(c *gin.Context) {

		collection := database.Collection("assistants")
		filer := bson.M{"setter": "admin"}
		var result bson.M
		if err := collection.FindOne(context.Background(), filer).Decode(&result); err != nil {
			log.Fatal(err)
		}

		c.JSON(http.StatusOK, gin.H{
			"result": result,
		})
	})

	r.POST("/login", func(c *gin.Context) {

		token, _ := utils.GenerateJwtToken("secret", "issuer", "audience", 60, 1, "admin", 1)

		response.OkWithData(gin.H{
			"token": token,
		}, c)
	})

	r.POST("/r", func(c *gin.Context) {
		var json interface{}

		c.ShouldBindJSON(&json)
		// if err := c.ShouldBindJSON(&json); err != nil {
		// 	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		// 	return
		// }

		response.OkWithData(gin.H{
			"message": json.(map[string]interface{})["name"],
		}, c)
	})

	r.POST("/", func(c *gin.Context) {

		c.BindJSON(&User{})

		collection := c.Value("db").(*mongo.Database).Collection("user")
		users := bson.M{"name": "ASliceOfBread", "password": "123456"}

		result, err := collection.InsertOne(context.Background(), users)

		if err != nil {
			log.Fatal(err)
		}

		// c.Value("db").(*mongo.Database).Collection("user").InsertOne(context.Background(), users)

		c.JSON(http.StatusOK, gin.H{
			"message": result,
		})

		// type User struct {
		// 	Name     string `json:"name"`
		// 	Password string `json:"password"`
		// }
		// var json User

		// if err := c.ShouldBindJSON(&json); err != nil {
		// 	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		// 	return
		// }
		// c.JSON(http.StatusOK, gin.H{
		// 	"message": json,
		// })
	})

	r.Run(":8080")

}
