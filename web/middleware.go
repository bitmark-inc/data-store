package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gopkg.in/macaroon.v2"
)

var (
	allowedCavearOps = map[string]string{
		"entity":    "=",
		"action":    "=",
		"resources": "in",
		"time":      "<",
	}
)

func (s *Server) CheckMacaroon() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		bearerTexts := strings.Split(auth, "Bearer ")
		if len(bearerTexts) != 2 {
			abortWithErrorMessage(c, http.StatusBadRequest, errorResponse{Message: "invalid Authorization header"})
			return
		}
		token := bearerTexts[1]

		var m macaroon.Macaroon
		data, err := macaroon.Base64Decode([]byte(token))
		if err != nil {
			abortWithErrorMessage(c, http.StatusBadRequest, errorResponse{Message: "macaroon not correctly encoded"})
			return
		}
		if err := m.UnmarshalBinary(data); err != nil {
			abortWithErrorMessage(c, http.StatusBadRequest, errorResponse{Message: "invalid macaroon"})
			return
		}

		caveats, err := m.VerifySignature(s.macaroonRootKey, nil)
		if err != nil {
			abortWithErrorMessage(c, http.StatusBadRequest, errorResponse{Message: "invalid macaroon signature"})
			return
		}

		for _, cav := range caveats {
			cond, op, arg, err := parseCaveat(cav)
			if err != nil {
				abortWithErrorMessage(c, http.StatusBadRequest, errorResponse{Message: "invalid caveat"})
				return
			}

			if op != allowedCavearOps[cond] {
				abortWithErrorMessage(c, http.StatusBadRequest, errorResponse{Message: "unknown operator in caveat"})
				return
			}

			switch cond {
			case "entity":
				c.Set("account_number", arg)
			case "action":
				switch c.Request.Method {
				case "POST", "PUT", "PATCH", "DELETE":
					if arg != "write" {
						forbidAccess(c, cav)
						return
					}
				case "GET", "HEAD":
					if arg != "read" {
						forbidAccess(c, cav)
						return
					}
				}
			case "resources":
				path := c.Request.URL.Path
				for _, param := range c.Params {
					path = strings.Replace(path, param.Value, "", -1)
				}
				path = strings.TrimRight(path, "/")
				parts := strings.Split(path, "/")
				targetResource := parts[len(parts)-1]
				allowedResources := strings.Split(arg, ",")
				valid := false
				for _, r := range allowedResources {
					if targetResource == r {
						valid = true
					}
				}
				if !valid {
					forbidAccess(c, cav)
					return
				}
			case "time":
				ts, err := strconv.ParseInt(arg, 10, 64)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"reason": "invalid ts in caveat"})
					return
				}

				expTime := time.Unix(ts, 0)
				if !time.Now().UTC().Before(expTime) {
					forbidAccess(c, cav)
					return
				}
			}
		}

		c.Next()
	}
}

func parseCaveat(cav string) (string, string, string, error) {
	if cav == "" {
		return "", "", "", fmt.Errorf("empty caveat")
	}
	parts := strings.Split(cav, " ")
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("invalid caveat format")
	}
	return parts[0], parts[1], parts[2], nil
}

func forbidAccess(c *gin.Context, cav string) {
	c.JSON(http.StatusForbidden, gin.H{"reason": fmt.Sprintf("caveat \"%s\" not satisfied", cav)})
	c.Abort()
}

func checkParticipant(filename string) gin.HandlerFunc {
	var participants = map[string]struct{}{}

	if filename == "" {
		panic("invalid participant file given")
	}

	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(&participants); err != nil {
		panic(err)
	}

	return func(c *gin.Context) {
		participantID := c.GetHeader("X-PARTICIPANT-ID")

		if _, ok := participants[participantID]; !ok {
			abortWithErrorMessage(c, http.StatusBadRequest, ErrParticipantNotAllowed)
			return
		}

		c.Next()
	}
}
