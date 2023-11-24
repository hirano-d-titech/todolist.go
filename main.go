package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"

	"todolist.go/db"
	"todolist.go/service"
)

const port = 8000

func main() {
	// initialize DB connection
	dsn := db.DefaultDSN(
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"))
	if err := db.Connect(dsn); err != nil {
		log.Fatal(err)
	}

	// initialize Gin engine
	engine := gin.Default()
	engine.LoadHTMLGlob("views/*.html")

	// prepare session
	store := cookie.NewStore([]byte("my-secret"))
	engine.Use(sessions.Sessions("user-session", store))

	// routing
	engine.Static("/assets", "./assets")
	engine.GET("/", service.Home)

	// ユーザ登録
	engine.GET("/user/new", service.NewUserForm)
	engine.POST("/user/new", service.RegisterUser)

	// ログイン
	engine.GET("/login", service.LoginForm)
	engine.POST("/login", service.Login)

	engine.GET("/user/logout", service.Logout)
	engine.GET("/user/delete", service.DeleteUser)

	// userGroup := engine.Group("/user")
	// userGroup.Use(service.LoginCheck)
	// {
	// 	userGroup.GET("/logout", service.Logout)
	// 	userGroup.GET("/delete", service.DeleteUser)
	// }

	engine.GET("/list", service.LoginCheck, service.TaskList)

	taskGroup := engine.Group("/task")
	taskGroup.Use(service.LoginCheck)
	{
		// タスク一覧
		taskGroup.GET("/:id", service.ShowTask) // ":id" is a parameter
		// タスクの新規登録
		taskGroup.GET("/new", service.NewTaskForm)
		taskGroup.POST("/new", service.RegisterTask)
		// 既存タスクの編集
		taskGroup.GET("/edit/:id", service.EditTaskForm)
		taskGroup.POST("/edit/:id", service.EditTask)
		// 既存タスクの削除
		taskGroup.GET("/delete/:id", service.DeleteTask)
	}

	// start server
	engine.Run(fmt.Sprintf(":%d", port))
}
