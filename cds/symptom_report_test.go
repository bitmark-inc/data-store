package cds

import (
	"testing"

	"github.com/bitmark-inc/data-store/store"
	"github.com/stretchr/testify/assert"
)

func TestTransformToSymptomDailyReports(t *testing.T) {
	lines := [][]string{
		{"symptoms", "2020-07-19", "2020-07-20", "2020-07-21"},
		{"Cough", "10", "7", "8"},
		{"Fatigue", "18", "28", "29"},
		{"numberCheckInOnce_Last3days", "NA", "NA", "1003"},
	}
	reports := transformToSymptomDailyReports(lines)
	assert.Equal(t, []store.SymptomDailyReport{
		{
			Date: "2020-07-19",
			Symptoms: []store.SymptomStats{
				{Name: "Cough", Count: 10},
				{Name: "Fatigue", Count: 18},
			},
			CheckinsNumPastThreeDays: 0,
		},
		{
			Date: "2020-07-20",
			Symptoms: []store.SymptomStats{
				{Name: "Cough", Count: 7},
				{Name: "Fatigue", Count: 28},
			},
			CheckinsNumPastThreeDays: 0,
		},
		{
			Date: "2020-07-21",
			Symptoms: []store.SymptomStats{
				{Name: "Cough", Count: 8},
				{Name: "Fatigue", Count: 29},
			},
			CheckinsNumPastThreeDays: 1003,
		}}, reports)
}
