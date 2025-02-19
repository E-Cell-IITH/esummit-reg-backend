package server

import (
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"

	"reg/internal/database"
)

type Server struct {
	port int
	*gin.Engine
}

func NewServer() *Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	database.New()

	server := &Server{
		port:   port,
		Engine: gin.Default(),
	}

	server.RegisterRoutes()

	return server
}
