package handlers

import (
	"fmt"
	"net/http"
	"site-controller-data-update-to-mysql/app/database"
	"site-controller-data-update-to-mysql/app/models"

	"context"

	"github.com/gin-gonic/gin"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func GetLatestTimestamp(c *gin.Context, db *database.Database) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	row, err := models.CSVUploadTransactions(
		qm.Select("*"),
		qm.OrderBy("timestamp DESC"),
	).One(ctx, db.DB)
	if err != nil {
		logging.Error(fmt.Sprintf("failed to get csv_upload_transaction: %v", err), nil)
		c.JSON(http.StatusInternalServerError, gin.H{"timestamp": nil})
		return
	}
	logging.Debug(fmt.Sprintf("csv_upload_transaction record: %p", row), nil)
	if row == nil {
		logging.Info("no csv information", nil)

		c.JSON(http.StatusOK, gin.H{"timestamp": nil})
		return
	}

	timestampStr := row.Timestamp.String
	timestampVal := fmt.Sprintf(`%s/%s/%s %s:%s:%s`, timestampStr[0:4], timestampStr[4:6], timestampStr[6:8], timestampStr[8:10], timestampStr[10:12], timestampStr[12:14])
	logging.Info(fmt.Sprintf("latest timestamp: %s", timestampVal), nil)
	c.JSON(http.StatusOK, gin.H{"timestamp": timestampVal})
	return
}
