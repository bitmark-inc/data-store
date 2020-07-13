package cds

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (cds *CDS) SetPOIRating() gin.HandlerFunc {
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

		err := cds.dataStorePool.Community().SetPOIRating(c, accountNumber, poiID, params.Ratings)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"result": "ok"})
	}
}

func (cds *CDS) GetPOISummarizedRating() gin.HandlerFunc {
	return func(c *gin.Context) {
		poiID := c.Param("poi_id")

		r, err := cds.dataStorePool.Community().GetPOISummarizedRating(c, poiID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, r)
	}
}
