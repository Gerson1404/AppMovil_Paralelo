package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	"go-hexagonal-api/internal/adapters/handlers"
	"go-hexagonal-api/internal/adapters/middleware"
	"go-hexagonal-api/internal/adapters/repositories"
	"go-hexagonal-api/internal/core/services"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

var ginLambda *ginadapter.GinLambda
var db *sql.DB 

func initDB() {
	dbURI := os.Getenv("DATABASE_URL")
	if dbURI == "" {
		dbURI = "postgres://postgres:123456@localhost:5432/hexagonal_db?sslmode=disable"
	}

	var err error
	db, err = sql.Open("postgres", dbURI)
	if err != nil {
		log.Fatalf("No se pudo conectar a la base de datos: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Base de datos inaccesible: %v", err)
	}

	createTableQuery := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		email VARCHAR(100) UNIQUE NOT NULL,
		password VARCHAR(255) NOT NULL,
		created_at TIMESTAMP NOT NULL
	);`
	if _, err := db.Exec(createTableQuery); err != nil {
		log.Fatalf("Error inicializando tablas: %v", err)
	}
}

func setupRouter() *gin.Engine {
	if db == nil {
		initDB()
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "super-secret-key-change-in-production"
	}

	userRepo := repositories.NewPostgresRepository(db)
	userService := services.NewUserService(userRepo, jwtSecret)
	userHandler := handlers.NewHTTPUserHandler(userService)

	router := gin.Default()

	// MODIFICACIÓN 3: Decirle a Gin que exponga los archivos que están en /tmp
	router.Static("/uploads", "/tmp")

	api := router.Group("/api")
	{
		api.POST("/auth/register", userHandler.Register)
		api.POST("/auth/login", userHandler.Login)
	}

	protected := router.Group("/api")
	protected.Use(middleware.JWTMiddleware(jwtSecret))
	{
		protected.GET("/users/:id", userHandler.GetByID)
		protected.PUT("/users/:id", userHandler.Update)
		protected.DELETE("/users/:id", userHandler.Delete)
		
		protected.POST("/upload", userHandler.UploadFile)
	}

	return router
}

func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return ginLambda.ProxyWithContext(ctx, req)
}

func main() {
	router := setupRouter()

	ginLambda = ginadapter.New(router)

	if os.Getenv("LAMBDA_TASK_ROOT") != "" {
		log.Println("Ejecutando en modo Serverless (AWS Lambda)...")
		lambda.Start(Handler)
	} else {
		log.Println("Ejecutando en modo Local (Docker/HTTP puerto 8080)...")
		
		defer db.Close() 
		
		if err := router.Run(":8080"); err != nil {
			log.Fatalf("Fallo al iniciar el servidor: %v", err)
		}
	}
}