package cds

import (
	"net/http"
	"strings"

	"github.com/bitmark-inc/data-store/store"
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

		err := cds.dataStorePool.Community("").SetPOIRating(c, accountNumber, poiID, params.Ratings)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"result": "ok"})
	}
}

func (cds *CDS) GetPOISummarizedRatings(c *gin.Context) {
	poiID := c.Param("poi_id")
	var result map[string]store.POISummarizedRating
	var err error
	if poiID != "" {
		result, err = cds.dataStorePool.Community("").GetPOISummarizedRatings(c, []string{poiID})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		var params struct {
			POIIDs string `form:"poi_ids" binding:"required"`
		}

		if err := c.BindQuery(&params); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		poiIDs := strings.Split(params.POIIDs, ",")

		result, err = cds.dataStorePool.Community("").GetPOISummarizedRatings(c, poiIDs)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	if len(result) == 0 {
		result = map[string]store.POISummarizedRating{}
	}
	c.JSON(http.StatusOK, result)
}
