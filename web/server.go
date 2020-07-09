package web

import (
	"context"
	"net/http"
	"time"

	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/bitmark-inc/data-store/store"
)

var log *logrus.Entry

func init() {
	log = logrus.WithField("prefix", "gin")
}

// Server to run a http server instance
type Server struct {
	server *http.Server

	dataStorePool store.DataStorePool

	location string
	rootKey  []byte
}

// NewServer new instance of server
func NewServer(mongoClient *mongo.Client) *Server {

	return &Server{
		location: "localhost",
		rootKey:  []byte("hello world"),

		dataStorePool: store.NewMongodbDataPool(mongoClient),
	}
}

// Run to run the server
func (s *Server) Run(addr string) error {
	s.server = &http.Server{
		Addr:    addr,
		Handler: s.setupRouter(),
	}

	return s.server.ListenAndServe()
}

func (s *Server) setupRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(sentrygin.New(sentrygin.Options{
		Repanic:         true,
		WaitForDelivery: false,
		Timeout:         10 * time.Second,
	}))

	apiRoute := r.Group("/api")
	apiRoute.POST("/register", s.Register)
	apiRoute.POST("/verify", s.Verify)

	poiRatingRoute := r.Group("/poi_rating")
	poiRatingRoute.GET("/:poi_id", s.GetPOIResource)
	poiRatingRoute.POST("", s.RatePOIResource)

	return r
}

// Shutdown to shutdown the server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func responseWithEncoding(c *gin.Context, code int, obj gin.H) {
	acceptEncoding := c.GetHeader("Accept-Encoding")
	switch acceptEncoding {
	default:
		c.JSON(code, obj)
	}
}

type ErrorResponse struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
}

func abortWithEncoding(c *gin.Context, code int, obj ErrorResponse, errors ...error) {
	for _, err := range errors {
		c.Error(err)
	}
	responseWithEncoding(c, code, gin.H{
		"error": obj,
	})
	c.Abort()
}
