package cds

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (p *CDS) ExportData(c *gin.Context) {
	accountNumber := c.GetString("account_number")

	data, err := p.dataStorePool.Community().ExportData(c, accountNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "application/octet-stream", data)
}
