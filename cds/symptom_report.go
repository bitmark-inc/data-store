package cds

import (
	"net/http"
	"sort"
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

type reportItem struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	Value        *int           `json:"value"`
	ChangeRate   *float64       `json:"change_rate"`
	Distribution map[string]int `json:"distribution"`
}

func (cds *CDS) GetSymptomReportItems() gin.HandlerFunc {
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

		current, err := cds.dataStorePool.Community().GetSymptomReportItems(c, start.Format("2006-01-02"), end.Format("2006-01-02"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		previous, err := cds.dataStorePool.Community().GetSymptomReportItems(c, prevStart.Format("2006-01-02"), prevEnd.Format("2006-01-02"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		results := gatherReportItemsWithDistribution(current, previous, false)
		items := getReportItemsForDisplay(results, func(symptomID string) string {
			return symptomID
		})
		c.JSON(http.StatusOK, gin.H{"report_items": items})
	}
}

func gatherReportItemsWithDistribution(currentBuckets, previousBuckets map[string][]store.Bucket, avg bool) map[string]*reportItem {
	items := make(map[string]*reportItem)
	for itemID, buckets := range currentBuckets {
		if len(buckets) == 0 {
			continue
		}
		// For each reported item shown in this period, assume it's not reported in the previous period,
		// So the change rate is 100 by default.
		// If it's also reported in the previous period, the rate will be adjusted accordingly.
		sum := 0
		distribution := make(map[string]int)
		for _, b := range buckets {
			sum += b.Value
			distribution[b.Name] = b.Value
		}
		value := sum
		if avg {
			value = sum / len(distribution)
		}
		changeRate := 100.0
		items[itemID] = &reportItem{ID: itemID, Value: &value, ChangeRate: &changeRate, Distribution: distribution}
	}
	for itemID, buckets := range previousBuckets {
		sum := 0
		for _, b := range buckets {
			sum += b.Value
		}

		if _, ok := items[itemID]; ok { // reported both in the current and previous periods
			changeRate := changeRate(float64(*items[itemID].Value), float64(sum))
			items[itemID].ChangeRate = &changeRate
		} else { // only reported in the previous period
			v := 0
			changeRate := -100.0
			items[itemID] = &reportItem{ID: itemID, Value: &v, ChangeRate: &changeRate}
		}
	}
	return items
}

func changeRate(new, old float64) float64 {
	if old == 0 {
		if new == 0 {
			return float64(0)
		} else {
			return float64(100)
		}
	}

	return (new - old) / old * 100
}

func getReportItemsForDisplay(entries map[string]*reportItem, getNameFunc func(string) string) []*reportItem {
	results := make([]*reportItem, 0)
	for entryID, entry := range entries {
		entry.Name = getNameFunc(entryID)
		results = append(results, entry)
	}
	sort.SliceStable(results, func(i, j int) bool {
		if *results[i].Value > *results[j].Value {
			return true
		}
		if *results[i].Value < *results[j].Value {
			return false
		}
		return results[i].Name < results[j].Name
	})
	return results
}
