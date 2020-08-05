package cds

import (
	"bufio"
	"encoding/csv"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/bitmark-inc/data-store/store"
)

const (
	onesignalAppID = "74f5ef01-1e4f-407e-a288-fa78fd552556"
)

var (
	notificationHeadingsNewReport = map[string]string{"en": "New trends data available"}
	notificationContentsNewReport = map[string]string{"en": "New trends data available"}
)

func (cds *CDS) AddSymptomDailyReports(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer f.Close()

	reader := csv.NewReader(bufio.NewReader(f))
	lines, err := reader.ReadAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	reports := transformToSymptomDailyReports(lines)
	if err := cds.dataStorePool.Community().AddSymptomDailyReports(c, reports); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// TODO: log error
	cds.notificationClient.NotifyActiveUsers(notificationHeadingsNewReport, notificationContentsNewReport)

	c.JSON(http.StatusOK, gin.H{"result": "ok"})
}

func transformToSymptomDailyReports(lines [][]string) []store.SymptomDailyReport {
	reports := make([]store.SymptomDailyReport, 0)
	for _, col := range lines[0][1:] {
		reports = append(reports, store.SymptomDailyReport{Date: col})
	}
	for i, row := range lines[1:] {
		if i == len(lines[1:])-1 {
			for i, cell := range row[1:] {
				if cell != "NA" {
					cnt, _ := strconv.Atoi(cell)
					reports[i].CheckinsNumPastThreeDays = cnt
				}
			}
			break
		}
		symptom := row[0]
		for i, cell := range row[1:] {
			cnt, _ := strconv.Atoi(cell)
			reports[i].Symptoms = append(reports[i].Symptoms, store.SymptomStats{Name: symptom, Count: cnt})
		}
	}
	return reports
}

type reportItemQueryParams struct {
	Days int64 `form:"days"`
}

type reportItem struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	Value        *int           `json:"value"`
	ChangeRate   *float64       `json:"change_rate"`
	Distribution map[string]int `json:"distribution"`
}

func (cds *CDS) GetSymptomReportItems(c *gin.Context) {
	var params reportItemQueryParams
	if err := c.Bind(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	days := int64(7)
	if params.Days != 0 {
		days = params.Days
	}

	latestReport, err := cds.dataStorePool.Community().FindLatestDailyReport(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	currEnd, err := time.Parse("2006-01-02", latestReport.Date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	prevEnd := currEnd.Add(-time.Hour * 24 * time.Duration(days))

	current, err := cds.dataStorePool.Community().GetSymptomReportItems(c, currEnd.Format("2006-01-02"), days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	previous, err := cds.dataStorePool.Community().GetSymptomReportItems(c, prevEnd.Format("2006-01-02"), days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	results := gatherReportItemsWithDistribution(current, previous, false)
	items := getReportItemsForDisplay(results, func(symptomID string) string {
		return symptomID
	})
	c.JSON(http.StatusOK, gin.H{
		"report_items":                 items,
		"checkins_num_past_three_days": latestReport.CheckinsNumPastThreeDays})
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
		if *results[i].Value == *results[j].Value {
			return results[i].Name < results[j].Name
		} else {
			return *results[i].Value > *results[j].Value
		}
	})
	return results
}
