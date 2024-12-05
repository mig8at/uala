package http

import (
	"fmt"
	"net/http"
	"strconv"
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
		authorized.GET("/users", s.paginate)
		authorized.GET("/users/:id", s.getById)
		authorized.POST("/users/:id/follow", s.follow)
		authorized.POST("/users/:id/unfollow", s.unfollow)
		authorized.GET("/users/:id/followers", s.followers)
		authorized.GET("/users/:id/following", s.following)
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

func (s *HTTPServer) paginate(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")

	// Convertir los parámetros a enteros
	pageInt, _ := strconv.Atoi(page)
	limitInt, _ := strconv.Atoi(limit)

	// Obtener los usuarios paginados
	users, err := s.userService.Paginate(c.Request.Context(), pageInt, limitInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Responder con los usuarios paginados
	c.JSON(http.StatusOK, users)
}

func (s *HTTPServer) getById(c *gin.Context) {
	id := c.Param("id")

	user, err := s.userService.GetById(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
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

func (s *HTTPServer) followers(c *gin.Context) {
	id := c.Param("id")
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")

	pageInt, _ := strconv.Atoi(page)
	limitInt, _ := strconv.Atoi(limit)

	followers, err := s.userService.Followers(c.Request.Context(), id, pageInt, limitInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, followers)
}

func (s *HTTPServer) following(c *gin.Context) {
	id := c.Param("id")
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")

	pageInt, _ := strconv.Atoi(page)
	limitInt, _ := strconv.Atoi(limit)

	following, err := s.userService.Following(c.Request.Context(), id, pageInt, limitInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, following)
}
