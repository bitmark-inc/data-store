package web

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gopkg.in/macaroon-bakery.v2/bakery/checkers"
	"gopkg.in/macaroon.v2"
)

func (s *Server) checkMacaroon() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		token := strings.Split(auth, "Bearer ")[1]

		var m macaroon.Macaroon
		data, err := hex.DecodeString(token)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"reason": "macaroon not hex-encoded"})
			return
		}
		if err := m.UnmarshalBinary(data); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"reason": err})
			return
		}

		// TODO: find the root key for this account number
		caveats, err := m.VerifySignature(s.rootKey, nil)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"reason": "invalid macaroon signature"})
			return
		}

		for _, cav := range caveats {
			cond, arg, err := checkers.ParseCaveat(cav)
			if err != nil {
				c.JSON(http.StatusForbidden, gin.H{"reason": err})
			}
			fmt.Println(cond, arg)

			switch cond {
			case "entity":
				c.Set("account_number", arg)
			case "action":
				switch c.Request.Method {
				case "POST", "PUT", "PATCH", "DELETE":
					if arg != "= write" {
						stop(c, cav)
						return
					}
				case "GET", "HEAD":
					if arg != "= read" {
						stop(c, cav)
						return
					}
				}
			case "resources":
				parts := strings.Split(c.Request.URL.Path, "/")
				targetResource := parts[len(parts)-1]

				if !strings.HasPrefix(arg, "in") {
					c.JSON(http.StatusBadRequest, gin.H{"reason": "unknown caveat"})
					return
				}
				allowedResources := strings.Split(strings.TrimLeft(arg, "in "), ",")

				valid := false
				for _, r := range allowedResources {
					if targetResource == r {
						valid = true
					}
				}
				if !valid {
					stop(c, cav)
					return
				}
			case "time":
				if !strings.HasPrefix(arg, "<") {
					c.JSON(http.StatusBadRequest, gin.H{"reason": "unknown caveat"})
					return
				}

				tsString := strings.TrimLeft(arg, "< ")
				ts, err := strconv.ParseInt(tsString, 10, 64)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"reason": "invalid ts in caveat"})
					return
				}

				expTime := time.Unix(ts, 0)
				if !time.Now().UTC().Before(expTime) {
					stop(c, cav)
				}
			}
		}
	}
}

func stop(c *gin.Context, cav string) {
	c.JSON(http.StatusForbidden, gin.H{"reason": fmt.Sprintf("caveat \"%s\" not satisfied", cav)})
	c.Abort()
}
