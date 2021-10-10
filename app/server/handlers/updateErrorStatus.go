package handlers

import (
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
		sugar.Errorf("failed to get csv_excution_error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"timestamp": nil}) // TODO 返却値をよく考えること
		return
	}
	sugar.Debugf("csv_excution_error record: %p", rows)
	if len(rows) == 0 {
		sugar.Info("no error csv")
		c.JSON(http.StatusOK, gin.H{"timestamp": nil})
		return
	}

	_, err = rows.UpdateAll(c, db.DB, models.M{"status": 1})
	if err != nil {
		sugar.Errorf("failed to get csv_excution_error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"timestamp": nil}) // TODO 返却値をよく考えること
		return
	}
	c.JSON(http.StatusOK, gin.H{"timestamp": nil})
}
