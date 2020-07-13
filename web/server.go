package web

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/bitmark-inc/bitmark-sdk-go/account"
)

// Server to run a http server instance
type Server struct {
	router *gin.Engine
	server *http.Server

	bitmarkAccount *account.AccountV2

	macaroonLocation string
	macaroonRootKey  []byte
}

// NewServer new instance of server
func NewServer(acct *account.AccountV2, endpoint string, macaroonRootKey []byte) *Server {
	r := gin.New()
	r.Use(gin.Recovery())

	return &Server{
		router:           r,
		bitmarkAccount:   acct,
		macaroonLocation: endpoint,
		macaroonRootKey:  macaroonRootKey,
	}
}

func (s *Server) Route(httpMethod, path string, handlers ...gin.HandlerFunc) {
	s.router.Handle(httpMethod, path, handlers...)
}

// Run to run the server
func (s *Server) Run(addr string) error {
	s.router.POST("/register", s.Register)
	s.router.GET("/information", s.Info)

	s.server = &http.Server{
		Addr:    addr,
		Handler: s.router,
	}

	return s.server.ListenAndServe()
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
