package pds

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (p *PDS) RatePOIResource() gin.HandlerFunc {
	return func(c *gin.Context) {
		accountNumber := c.GetString("account_number")
		poiID := c.Param("poi_id")

		var params struct {
			Ratings map[string]float64 `json:"ratings"`
		}

		if err := c.Bind(&params); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if params.Ratings == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no rating provided"})
			return
		}

		err := p.dataStorePool.Account(accountNumber).SetPOIRating(c, poiID, params.Ratings)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"result": "ok"})
	}
}

func (p *PDS) GetPOIResource() gin.HandlerFunc {
	return func(c *gin.Context) {
		accountNumber := c.GetString("account_number")
		poiID := c.Param("poi_id")

		r, err := p.dataStorePool.Account(accountNumber).GetPOIRating(c, poiID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if r == nil {
			r = map[string]float64{}
		}

		c.JSON(http.StatusOK, r)
	}
}
