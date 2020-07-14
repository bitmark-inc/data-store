package cds

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/bitmark-inc/data-store/store"
)

func (cds *CDS) AddSymptomDailyReports() gin.HandlerFunc {
	return func(c *gin.Context) {
		var body struct {
			Reports []*store.SymptomDailyReport `json:"reports"`
		}

		if err := c.Bind(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := cds.dataStorePool.Community().AddSymptomDailyReports(c, body.Reports); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"result": "ok"})
	}
}

type reportItemQueryParams struct {
	Start string `form:"start"`
	End   string `form:"end"`
}

func (cds *CDS) GetSymptomReports() gin.HandlerFunc {
	return func(c *gin.Context) {
		var params reportItemQueryParams
		if err := c.Bind(&params); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		start, err := time.Parse(time.RFC3339, params.Start)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		end, err := time.Parse(time.RFC3339, params.End)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		gap := start.Sub(end)
		prevStart := start.Add(gap)
		prevEnd := start

		current, err := cds.dataStorePool.Community().GetSymptomReports(c, start.Format("2006-01-02"), end.Format("2006-01-02"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		previous, err := cds.dataStorePool.Community().GetSymptomReports(c, prevStart.Format("2006-01-02"), prevEnd.Format("2006-01-02"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"current_period":  current,
			"previous_period": previous,
		})
	}
}
