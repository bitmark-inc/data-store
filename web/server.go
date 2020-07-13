package web

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/bitmark-inc/bitmark-sdk-go/account"
	"github.com/bitmark-inc/data-store/store"
)

// Server to run a http server instance
type Server struct {
	server *http.Server

	dataStorePool store.DataStorePool

	bitmarkAccount *account.AccountV2

	macaroonLocation string
	macaroonRootKey  []byte
}

// NewServer new instance of server
func NewServer(mongoClient *mongo.Client, acct *account.AccountV2, endpoint string, macaroonRootKey []byte) *Server {
	return &Server{
		dataStorePool:    store.NewMongodbDataPool(mongoClient),
		bitmarkAccount:   acct,
		macaroonLocation: endpoint,
		macaroonRootKey:  macaroonRootKey,
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

	apiRoute := r.Group("/api")
	apiRoute.POST("/register", s.Register)

	poiRatingRoute := r.Group("/poi_rating")
	poiRatingRoute.Use(s.checkMacaroon())
	poiRatingRoute.GET("/:poi_id", s.GetPOIRating)
	poiRatingRoute.PUT("/:poi_id", s.SetPOIRating)

	return r
}

// Shutdown terminates the web server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

type errorResponse struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
}

func abortWithErrorMessage(c *gin.Context, code int, resp errorResponse, errors ...error) {
	for _, err := range errors {
		c.Error(err)
	}
	c.JSON(code, resp)
	c.Abort()
}
