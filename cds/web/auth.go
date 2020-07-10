package web

import (
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gopkg.in/macaroon.v2"
)

func (s *Server) Register(c *gin.Context) {
	var req struct {
		Timestamp string `json:"timestamp"`
		Signature string `json:"signature"`
		Requester string `json:"requester"`
	}

	if err := c.BindJSON(&req); err != nil {
		abortWithEncoding(c, http.StatusBadRequest, ErrorResponse{
			Message: err.Error(),
		}, err)
		return
	}

	// sig, err := hex.DecodeString(req.Signature)
	// if err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"reason": "signature not hex-encoded"})
	// 	return
	// }

	// if err := account.Verify(req.Requester, []byte(req.Timestamp), sig); err != nil {
	// 	c.JSON(http.StatusForbidden, gin.H{"reason": "invalid signature"})
	// 	return
	// }

	// mint macaroons
	// uuid, err := uuid.NewRandom()
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"reason": err})
	// 	return
	// }
	rootMacaroon, err := macaroon.New(s.rootKey, []byte("ace6795b-1957-43d5-8148-81bd2190f1a3"), s.location, macaroon.V1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err})
		return
	}
	readMacaroon, err := createMacaroon(rootMacaroon, "read", req.Requester)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err})
		return
	}
	writeMacaroon, err := createMacaroon(rootMacaroon, "write", req.Requester)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err})
		return
	}
	rm, err := readMacaroon.MarshalBinary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err})
		return
	}
	wm, err := writeMacaroon.MarshalBinary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err})
		return
	}
	fmt.Println(string(readMacaroon.Id()), hex.EncodeToString(rm))
	c.JSON(http.StatusOK, gin.H{
		"r": hex.EncodeToString(rm),
		"w": hex.EncodeToString(wm),
	})
}

func (s *Server) Verify(c *gin.Context) {
	var req struct {
		Macaroon string `json:"macaroon"`
	}

	if err := c.BindJSON(&req); err != nil {
		abortWithEncoding(c, http.StatusBadRequest, ErrorResponse{
			Message: err.Error(),
		}, err)
		return
	}

	var m macaroon.Macaroon
	data, _ := hex.DecodeString(req.Macaroon)
	if err := m.UnmarshalBinary(data); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"reason": "macaroon not hex-encoded"})
		return
	}

	caveats, err := m.VerifySignature(s.rootKey, nil)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"reason": "invalid signature"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"caveats": caveats,
	})
}

func createMacaroon(m *macaroon.Macaroon, op, accountNumber string) (*macaroon.Macaroon, error) {
	cloned := m.Clone()
	err := cloned.AddFirstPartyCaveat([]byte(fmt.Sprintf("%s @ %s", op, accountNumber)))
	return cloned, err
}
