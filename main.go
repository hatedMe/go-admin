package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	response "example.com/mod/core"
	"example.com/mod/middleware"
	"example.com/mod/utils"
)

type Config struct {
	Redis struct {
		DB       int    `mapstructure:"db" json:"db" yaml:"db"`                   // redis的哪个数据库
		Addr     string `mapstructure:"addr" json:"addr" yaml:"addr"`             // 服务器地址:端口
		Password string `mapstructure:"password" json:"password" yaml:"password"` // 密码
	} `mapstructure:"redis" json:"redis" yaml:"redis"`

	MongoDB struct {
		Host     string `mapstructure:"host" json:"host" yaml:"host"`             // 数据库地址
		Port     string `mapstructure:"port" json:"port" yaml:"port"`             // 数据库端口
		Dbname   string `mapstructure:"dbname" json:"dbname" yaml:"dbname"`       // 数据库名
		Username string `mapstructure:"username" json:"username" yaml:"username"` // 数据库用户名
		Password string `mapstructure:"password" json:"password" yaml:"password"` // 数据库密码
	} `mapstructure:"mongodb" json:"mongodb" yaml:"mongodb"`
}

var (
	clientOptions *options.ClientOptions
	database      *mongo.Database
	env           = flag.String("env", "dev", "the environment to use (dev, test, prod)")
	config        Config
)

func main() {

	flag.Parse()
	fmt.Println(*env, "env---")

	// 使用 Viper 来管理配置
	v := viper.New()
	v.SetConfigFile("./config/config." + *env + ".yaml")
	v.SetConfigType("yaml")
	// 设置配置文件路径
	v.AddConfigPath(".")
	err := v.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	v.SetConfigName("./config/config." + *env + ".local.yaml")
	v.MergeInConfig()
	v.Unmarshal(&config)

	fmt.Println(config.Redis.Addr, config.Redis.Password, config.Redis.DB)

	r := gin.New()

	clientOptions := options.Client().ApplyURI("mongodb://" + config.MongoDB.Username + ":" + config.MongoDB.Password + "@" + config.MongoDB.Host + ":" + config.MongoDB.Port + "/" + config.MongoDB.Dbname)

	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	redisClient := redis.NewClient(&redis.Options{
		DB:       config.Redis.DB,
		Addr:     config.Redis.Addr,
		Password: config.Redis.Password,
	})

	database := client.Database("blog")

	r.Use(func(c *gin.Context) {
		c.Set("db", database)
		c.Next()
	})

	api := r.Group("/api")

	admin := api.Group("/admin")
	{
		article := admin.Group("/article", middleware.Auth())
		{
			article.POST("/add", func(c *gin.Context) {
				type Article struct {
					Id         string    `json:"id" bson:"id,omitempty"`
					CreateTime time.Time `json:"createTime" bson:"createTime"`
					UpdateTime time.Time `json:"updateTime" bson:"updateTime"`
					Title      string    `json:"title"`
					Content    string    `json:"content"`
					Summary    string    `json:"summary"`
					Labels     []string  `json:"labels"`
					Category   []string  `json:"category"`
				}
				var article Article
				c.ShouldBindJSON(&article)
				article.Id = utils.CreateRandId()
				article.CreateTime = time.Now()
				article.UpdateTime = time.Now()

				collection := c.Value("db").(*mongo.Database).Collection("article")
				_, err := collection.InsertOne(context.Background(), article)
				if err != nil {
					response.FailWithMessage("添加失败", c)
				}
				response.OkWithData(article, c)
			})

			article.POST("/update", func(c *gin.Context) {
				type Article struct {
					Id         string    `json:"id" bson:"id,omitempty"`
					CreateTime time.Time `json:"createTime" bson:"createTime"`
					UpdateTime time.Time `json:"updateTime" bson:"updateTime"`
					Title      string    `json:"title"`
					Content    string    `json:"content"`
					Summary    string    `json:"summary"`
					Labels     []string  `json:"labels"`
					Category   []string  `json:"category"`
				}
				var article Article
				c.ShouldBindJSON(&article)

				// 定义要修改的文档的查询条件
				filter := bson.M{"id": article.Id}

				// 定义要更新的字段和新值
				update := bson.M{"$set": bson.M{
					"title":      article.Title,
					"content":    article.Content,
					"summary":    article.Summary,
					"labels":     article.Labels,
					"category":   article.Category,
					"updateTime": time.Now(),
				}}

				collection := c.Value("db").(*mongo.Database).Collection("article")
				result, err := collection.UpdateOne(context.Background(), filter, update)
				if err != nil || result == nil || result.MatchedCount == 0 {
					response.FailWithMessage("修改失败", c)
					return
				}
				response.OkWithMessage("修改成功", c)
			})

			article.GET("/getAll", func(c *gin.Context) {
				collection := c.Value("db").(*mongo.Database).Collection("article")
				var result []bson.M

				cursor, _ := collection.Find(context.Background(), bson.M{}, options.Find().SetProjection(bson.M{"_id": 0}))
				if err != nil {
					response.FailWithMessage("查询失败", c)
					return
				}
				cursor.All(context.Background(), &result)
				response.OkWithData(gin.H{
					"lists": result,
					"total": len(result),
				}, c)
			})

			article.GET("/getItemById", func(c *gin.Context) {
				userID := c.Query("id")
				id, _ := primitive.ObjectIDFromHex(userID)
				// 定义要修改的文档的查询条件
				filter := bson.M{"_id": id}
				collection := c.Value("db").(*mongo.Database).Collection("article")

				var result bson.M
				err := collection.FindOne(context.Background(), filter, options.FindOne().SetProjection(bson.M{"_id": 0})).Decode(&result)
				if err != nil {
					response.FailWithMessage("查询失败", c)
					return
				}
				response.OkWithData(result, c)

			})

			// article.GET("/delete", func(c *gin.Context) {

			// })
		}

		user := admin.Group("/user")
		{
			user.POST("/register", func(c *gin.Context) {
				type User struct {
					Id         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
					CreateTime time.Time          `json:"createTime" bson:"createTime"`
					UpdateTime time.Time          `json:"updateTime" bson:"updateTime"`
					RoleId     int                `json:"roleId" bson:"roleId"`
					Name       string             `json:"name"`
					Password   string             `json:"password"`
				}
				var user User
				user.CreateTime = time.Now()
				user.UpdateTime = time.Now()
				c.ShouldBindJSON(&user)

				user.Password = utils.Md5(user.Password)

				collection := c.Value("db").(*mongo.Database).Collection("user")
				_, err := collection.InsertOne(context.Background(), user)
				if err != nil {
					response.FailWithMessage("添加失败", c)
				}
				response.OkWithData(nil, c)
			})
			user.POST("/login", func(c *gin.Context) {
				type User struct {
					Name     string `json:"name"`
					Password string `json:"password"`
				}
				var user User
				c.ShouldBindJSON(&user)
				user.Password = utils.Md5(user.Password)

				collection := c.Value("db").(*mongo.Database).Collection("user")
				fmt.Println("user")
				var result bson.M

				if err := collection.FindOne(context.Background(), bson.M{"name": user.Name, "password": user.Password}).Decode(&result); err != nil {
					response.FailWithMessage("登录失败", c)
					return
				}

				token, _ := utils.GenerateJwtToken("secret", "issuer", "audience", int64(time.Hour)*24, result["_id"].(primitive.ObjectID).String(), result["name"].(string), 1)

				redisClient.Set(context.Background(), "ADMIN_TOKEN_"+strings.ToUpper(utils.Md5(token)), token, 7*time.Hour*24)

				fmt.Println(result)
				fmt.Println(result["_id"].(primitive.ObjectID).String())

				response.OkWithData(gin.H{
					"token":  token,
					"name":   result["name"],
					"roleId": result["roleId"],
				}, c)

			})
		}

	}

	front := api.Group("/front")
	{
		front.GET("/getArticleAll", func(c *gin.Context) {
			collection := c.Value("db").(*mongo.Database).Collection("article")
			var result []bson.M

			cursor, _ := collection.Find(context.Background(), bson.M{}, options.Find().SetProjection(bson.M{"_id": 0}))
			if err != nil {
				response.FailWithMessage("查询失败", c)
			}
			cursor.All(context.Background(), &result)
			response.OkWithData(gin.H{
				"list":  result,
				"total": len(result),
			}, c)
		})
	}

	r.Run(":8986")

}
