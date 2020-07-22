package pds

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (p *PDS) ExportData(c *gin.Context) {
	accountNumber := c.GetString("account_number")

	data, err := p.dataStorePool.Account(accountNumber).ExportData(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "application/octet-stream", data)
}

func (p *PDS) DeleteData(c *gin.Context) {
	accountNumber := c.GetString("account_number")

	err := p.dataStorePool.Account(accountNumber).DeleteData(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": "ok"})
}
