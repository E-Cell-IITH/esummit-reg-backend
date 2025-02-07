package server

import (
	"net/http"
	"os"
	"reg/internal/controllers"
	"reg/internal/database"
	paymentgateway "reg/internal/payment_gateway"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterRoutes() http.Handler {
	allowedOrigins := os.Getenv("ALLOW_DOMAINS")
	if allowedOrigins == "" {
		allowedOrigins = "http://localhost:3000"
	}
	origins := strings.Split(allowedOrigins, ",")

	s.Use(cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true, // Enable cookies/auth
	}))
	s.Use(AuthMiddleware())

	s.GET("/", s.HelloWorldHandler)

	s.GET("/health", s.healthHandler)
	s.POST("/register", controllers.RegisterHandler)
	s.POST("/update-startup-sheet", controllers.PostDataInGSheet)

	// E-Summit-2025

	signup := s.Group("/signup")
	{
		signup.POST("", controllers.RegisterUserHandler)
		signup.POST("/otp/send", controllers.SendOtpSignUP)
		signup.POST("/otp/verify", controllers.VerifyOtpSignUP)
	}

	signin := s.Group("/signin")
	{
		signin.POST("", controllers.VerifyOtpSignIN)
		signin.POST("/otp/send", controllers.SendOtpSignIN)
	}

	s.GET("/me", controllers.GetUserHandler)
	s.GET("/logout", controllers.LogoutHandler)

	s.POST("/paymentInitiate", paymentgateway.CreateOrder)
	s.POST("/transactionID", paymentgateway.PushTransactionIds)
	s.POST("/applyCoupon", paymentgateway.HandleCouponVerifications)

	admin := s.Group("/admin")
	{
		admin.POST("/transactionID", paymentgateway.AddSuccessfulTxnIds)
	}

	s.GET(("/passes"), controllers.SendPassesHandler)

	return s
}

func (s *Server) HelloWorldHandler(c *gin.Context) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"

	c.JSON(http.StatusOK, resp)
}

func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, database.Health())
}
