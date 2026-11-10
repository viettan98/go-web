package main

import (
	"log"
	"web/internal/database"
	"web/internal/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Initialize MySQL database connection (Thay đổi mật khẩu root của bạn ở đây nếu cần)
	dsn := "root:Tan0819151575@tcp(127.0.0.1:3306)/gobook_db?parseTime=true"
	database.InitMySQL(dsn)

	// 2 & 3. Initialize AppHandler (Tự động khởi tạo các kho NoSQL độc lập và kết nối DB)
	appHandler := handlers.NewAppHandler()

	// 4. Setup Gin router
	router := gin.Default()

	// Load HTML templates from the "templates" folder
	router.Static("/static", "./static")
	router.LoadHTMLGlob("templates/*")

	// 5. Register application routes (both API and Web UI)
	appHandler.RegisterRoutes(router)

	// 5. Start the web server
	log.Println("GoBook server is running on http://127.0.0.1:18085")
	if err := router.Run("127.0.0.1:18085"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}