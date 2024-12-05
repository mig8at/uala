package http

import (
	"fmt"
	"net/http"
	"user_service/internal/application/dto"
	"user_service/internal/interfaces"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type HTTPServer struct {
	engine      *gin.Engine
	validate    *validator.Validate
	userService interfaces.UserService
}

func NewHTTPServer(engine *gin.Engine, userService interfaces.UserService, validate *validator.Validate) *HTTPServer {
	server := &HTTPServer{
		engine:      engine,
		validate:    validate,
		userService: userService,
	}
	server.registerRoutes()
	return server
}

func (s *HTTPServer) Run(port string) {
	if err := s.engine.Run(port); err != nil {
		panic(err)
	}
}

func (s *HTTPServer) registerRoutes() {
	s.engine.POST("/users", s.create)
	authorized := s.engine.Group("/", AuthMiddleware())
	{
		authorized.POST("/users/:id/follow", s.follow)
		authorized.POST("/users/:id/unfollow", s.unfollow)
	}
}

func (s *HTTPServer) create(c *gin.Context) {
	var user dto.CreateUser

	// Intentar vincular el cuerpo JSON al DTO
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validar los datos con el validador
	if err := s.validate.Struct(user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Error de validación: %s", err.Error())})
		return
	}

	// Crear el usuario usando el servicio
	createdUser, err := s.userService.Create(c.Request.Context(), &user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Responder con el usuario creado
	c.JSON(http.StatusCreated, createdUser)
}

func (s *HTTPServer) follow(c *gin.Context) {
	userID, _ := c.Get("userID")
	id := userID.(string)
	followerID := c.Param("id")

	err := s.userService.Follow(c.Request.Context(), id, followerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Usuario seguido correctamente."})
}

func (s *HTTPServer) unfollow(c *gin.Context) {
	userID, _ := c.Get("userID")
	id := userID.(string)
	followerID := c.Param("id")

	err := s.userService.Unfollow(c.Request.Context(), id, followerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Usuario dejado de seguir correctamente."})
}
