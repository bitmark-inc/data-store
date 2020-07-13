package web

import (
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) Info(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"server": gin.H{
			"bitmark_account_number": s.bitmarkAccount.AccountNumber(),
			"enc_pub_key":            hex.EncodeToString(s.bitmarkAccount.EncrKey.PublicKeyBytes()),
		},
	})
}
