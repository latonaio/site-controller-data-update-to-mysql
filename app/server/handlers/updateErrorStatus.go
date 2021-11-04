package handlers

import (
	"fmt"
	"net/http"
	"site-controller-data-update-to-mysql/app/database"
	"site-controller-data-update-to-mysql/app/models"

	"github.com/gin-gonic/gin"
)

func UpdateErrorStatus(c *gin.Context, db *database.Database) {
	rows, err := models.CSVExecutionErrors(
		models.CSVExecutionErrorWhere.Status.EQ(0),
	).All(c, db.DB)
	if err != nil {
		logging.Error(fmt.Sprintf("failed to get csv_excution_error: %v", err), nil)
		c.JSON(http.StatusInternalServerError, gin.H{"timestamp": nil})
		return
	}
	logging.Debug(fmt.Sprintf("csv_excution_error record: %p", rows), nil)

	if len(rows) == 0 {
		logging.Info("no error csv", nil)
		c.JSON(http.StatusOK, gin.H{"timestamp": nil})
		return
	}

	_, err = rows.UpdateAll(c, db.DB, models.M{"status": 1})
	if err != nil {
		logging.Error(fmt.Sprintf("failed to get csv_excution_error: %v", err), nil)
		c.JSON(http.StatusInternalServerError, gin.H{"timestamp": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"timestamp": nil})
}
