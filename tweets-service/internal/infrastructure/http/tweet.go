package http

import (
	"fmt"
	"net/http"
	"tweet-service/internal/application/dto"
	"tweet-service/internal/interfaces"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type HTTPServer struct {
	engine       *gin.Engine
	validate     *validator.Validate
	tweetservice interfaces.Tweetservice
}

func NewHTTPServer(engine *gin.Engine, tweetservice interfaces.Tweetservice, validate *validator.Validate) *HTTPServer {
	server := &HTTPServer{
		engine:       engine,
		validate:     validate,
		tweetservice: tweetservice,
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
		authorized.POST("/tweets", s.create)
		authorized.DELETE("/tweets/:id", s.delete)

	}
}

func (s *HTTPServer) create(c *gin.Context) {
	var tweet dto.CreateTweet

	if err := c.ShouldBindJSON(&tweet); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.validate.Struct(tweet); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Error de validación: %s", err.Error())})
		return
	}

	createdtweet, err := s.tweetservice.Create(c.Request.Context(), &tweet)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, createdtweet)
}

func (s *HTTPServer) delete(c *gin.Context) {
	id := c.Param("id")

	if err := s.tweetservice.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tweet eliminado correctamente"})
}
