package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) RatePOIResource(c *gin.Context) {
	accountNumber := c.GetString("account_number")
	poiID := c.Param("poi_id")

	var params struct {
		Ratings map[string]float64 `json:"ratings"`
	}

	if err := c.Bind(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := s.dataStorePool.Account(accountNumber).RatePOIResource(c, poiID, params.Ratings)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": "ok"})
}

func (s *Server) GetPOIResource(c *gin.Context) {
	accountNumber := c.GetString("account_number")
	poiID := c.Param("poi_id")

	r, err := s.dataStorePool.Account(accountNumber).GetPOIResource(c, poiID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, r)
}
