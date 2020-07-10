package web

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

	"github.com/bitmark-inc/bitmark-sdk-go/account"
	"github.com/gin-gonic/gin"
	"gopkg.in/macaroon.v2"
)

func (s *Server) Register(c *gin.Context) {
	var req struct {
		Timestamp string `json:"timestamp"`
		Signature string `json:"signature"`
		Requester string `json:"requester"`
		EncKey    string `json:"encryption_public_key"`
	}

	if err := c.BindJSON(&req); err != nil {
		abortWithErrorMessage(c, http.StatusBadRequest, errorResponse{Message: err.Error()}, err)
		return
	}
	sig, err := hex.DecodeString(req.Signature)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "signature not hex-encoded"})
		return
	}
	msg := strings.Join([]string{req.EncKey, req.Timestamp}, "|")
	if err := account.Verify(req.Requester, []byte(msg), sig); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"reason": "invalid signature"})
		return
	}

	recipientPublicKey, err := hex.DecodeString(req.EncKey)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "encryption_public_key not hex-encoded"})
		return
	}

	// mint macaroons
	rootMacaroon, err := macaroon.New(s.macaroonRootKey, []byte(req.Requester), s.macaroonLocation, macaroon.V1)
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

	encryptedReadMacaroon, err := encryptMacaroon(readMacaroon, recipientPublicKey, s.bitmarkAccount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err})
		return
	}
	encryptedWriteMacaroon, err := encryptMacaroon(writeMacaroon, recipientPublicKey, s.bitmarkAccount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"r": encryptedReadMacaroon,
		"w": encryptedWriteMacaroon,
	})
}

func createMacaroon(m *macaroon.Macaroon, op, accountNumber string) (*macaroon.Macaroon, error) {
	cloned := m.Clone()
	cloned.AddFirstPartyCaveat([]byte(fmt.Sprintf("entity = %s", accountNumber)))
	cloned.AddFirstPartyCaveat([]byte(fmt.Sprintf("action = %s", op)))
	return cloned, nil
}

func encryptMacaroon(macaroon *macaroon.Macaroon, recipientPublicKey []byte, senderAccount *account.AccountV2) (string, error) {
	data, err := macaroon.MarshalBinary()
	if err != nil {
		return "", err
	}

	ciphertext, err := senderAccount.EncrKey.Encrypt(data, recipientPublicKey)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(ciphertext), nil
}
