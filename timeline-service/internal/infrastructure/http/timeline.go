package http

import (
	"net/http"
	"strconv"
	"timeline-service/internal/interfaces"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type HTTPServer struct {
	engine   *gin.Engine
	validate *validator.Validate
	service  interfaces.Service
}

func NewHTTPServer(engine *gin.Engine, service interfaces.Service, validate *validator.Validate) *HTTPServer {
	server := &HTTPServer{
		engine:   engine,
		validate: validate,
		service:  service,
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
	authorized := s.engine.Group("/", AuthMiddleware())
	{
		authorized.GET("/timeline", s.paginate)

	}
}

func (s *HTTPServer) paginate(c *gin.Context) {
	id := c.GetString("userID")
	limit := c.DefaultQuery("limit", "10")
	offset := c.DefaultQuery("offset", "0")

	limitInt, _ := strconv.Atoi(limit)
	offsetInt, _ := strconv.Atoi(offset)

	// Paginar los usuarios usando el servicio
	tweets, err := s.service.Paginate(c.Request.Context(), id, limitInt, offsetInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Responder con los usuarios paginados
	c.JSON(http.StatusOK, tweets)
}
