package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) addCheckinsData(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}
