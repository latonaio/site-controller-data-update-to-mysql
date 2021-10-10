package handlers

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"time"
	"site-controller-data-update-to-mysql/app/database"
	"site-controller-data-update-to-mysql/app/file"
	"site-controller-data-update-to-mysql/app/models"

	"github.com/gin-gonic/gin"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"go.uber.org/zap"
)

const (
	CONSARVATION_PATH string = "/var/lib/aion/Data"
)

type errors struct {
	LineNumber          int    `json:"line_number"`
	CustomerName        string `json:"customer_name"`
	CustomerPhoneNumber string `json:"customer_phone_number"`
}

type SCHandler struct {
	db  *database.Database
	log *zap.SugaredLogger
}

func NewSCHandler(db *database.Database, logger *zap.SugaredLogger) *SCHandler {
	return &SCHandler{
		db:  db,
		log: logger,
	}
}

func (h *SCHandler) GetAuthCSV(c *gin.Context) {
	timestamp := c.Param("timestamp")
	rows, err := h.db.GetCsvUploadTransactionByTimeStamp(c.Request.Context(), timestamp)
	if err != nil {
		h.log.Errorf("database error: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "before"})
		return
	}
	if len(rows) == 1 {
		row := rows[0]
		c.JSON(http.StatusOK, gin.H{"timestamp": row.Timestamp.String, "status": row.Status.String, "path": row.Path.String})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"status": "before"})
	return
}

func (h *SCHandler) CSVError(c *gin.Context) {
	// ステータスが未解決（0)のCSVエラー情報を取得する
	rows, err := h.db.GetCsvExecutionErrorsWithCsvUploadTransactionByStatus(c.Request.Context(), 0)
	if err != nil {
		sugar.Errorf("cannot get csv error rows: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": []errors{}})
		return
	}
	if len(rows) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"errors": []errors{},
		})
		return
	}

	row := rows[0]
	res := map[string]interface{}{
		"file_name": row.R.CSV.FileName.String,
		"errors": []errors{
			{
				LineNumber:          row.LineNumber,
				CustomerName:        row.CustomerName.String,
				CustomerPhoneNumber: row.CustomerPhoneNumber.String,
			},
		},
	}
	c.JSON(http.StatusOK, res)
	return
}

func (h *SCHandler) UpdateErrorStatus(c *gin.Context) {
	// ステータスが未解決（0)のレコードを取得する
	rows, err := h.db.GetCsvExecutionErrorsByStatus(c.Request.Context(), 0)
	if err != nil {
		sugar.Errorf("failed to get csv_excution_error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"timestamp": nil}) // TODO 返却値をよく考えること
		return
	}
	h.log.Debugf("csv_excution_error record: %p", rows)
	if len(rows) == 0 {
		h.log.Debug("no error csv")
		c.JSON(http.StatusOK, gin.H{"timestamp": nil})
		return
	}

	_, err = rows.UpdateAll(c, h.db.DB, models.M{"status": 1})
	if err != nil {
		sugar.Errorf("failed to get csv_excution_error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"timestamp": nil}) // TODO 返却値をよく考えること
		return
	}
	c.JSON(http.StatusOK, gin.H{"timestamp": nil})
}

func (h *SCHandler) CreateCSV(c *gin.Context) {
	timestamp := c.Param("timestamp")
	siteControllerName := c.Query("SC")
	if siteControllerName == "" {
		h.log.Errorf("failed to get site controller name")
		c.String(http.StatusBadRequest, "BAD REQUEST")
		return
	}
	ctx := c.Request.Context()

	// リクエストの情報を出力
	body, err := httputil.DumpRequest(c.Request, true)
	if err != nil {
		h.log.Errorf("failed to bind request body: %v", err)
		c.String(http.StatusInternalServerError, "INTERNAL SERVER ERROR")
		return
	}
	h.log.Debugf("req body: %v", body)

	// "file"というフィールド名に一致するファイルが出力される
	formFile, _, err := c.Request.FormFile("file")
	if err != nil {
		h.log.Errorf("failed to get file: %v", err)
		c.String(http.StatusInternalServerError, "INTERNAL SERVER ERROR")
		return
	}
	defer formFile.Close()

	// ファイル名をトリム
	defaultFileName := c.Request.FormValue("defaultFileName")
	trimmedDefaultFileName := strings.Trim(defaultFileName, ".csv")

	// csvファイルの情報を取得するために一度ファイルをローカル(コンテナ内）に保存する
	//データを保存するファイルを開く
	filePath := fmt.Sprintf("%v/%v_%v.csv", CONSARVATION_PATH, trimmedDefaultFileName, timestamp)
	saveFile, err := os.Create(filePath)
	if err != nil {
		h.log.Errorf("failed to save file: %v", err)
		c.String(http.StatusInternalServerError, "INTERNAL SERVER ERROR")
		return
	}
	defer saveFile.Close()

	// ファイルにデータを書き込む
	_, err = io.Copy(saveFile, formFile)
	if err != nil {
		sugar.Errorf("failed to copy file: %+v", err)
		c.String(http.StatusInternalServerError, "INTERNAL SERVER ERROR")
		return
	}

	// ファイルのデータをDBに突っ込む
	fileInfo, err := saveFile.Stat()
	if err != nil {
		sugar.Errorf("failed to get file info: %+v", err)
		c.String(http.StatusInternalServerError, "INTERNAL SERVER ERROR")
		return
	}
	file := file.File{
		Name:        fileInfo.Name(),
		CreatedTime: fileInfo.ModTime(),
	}
	model, err := h.db.CreateCsvUploadTransaction(ctx, file.Name, time.Time{}, timestamp, filePath)
	if err != nil {
		sugar.Errorf("failed to insert record to database: %+v", err)
		c.String(http.StatusInternalServerError, "INTERNAL SERVER ERROR")
		return
	}

	if err := h.db.RegisterCSVDataToDB(ctx, file, CONSARVATION_PATH, model.ID, siteControllerName); err != nil {
		sugar.Error(err)
		c.String(http.StatusInternalServerError, "INTERNAL SERVER ERROR")
		return
	}

	c.String(http.StatusOK, "ok")
}

func (h *SCHandler) GetLatestTimestamp(c *gin.Context) {
	ctx := c.Request.Context()

	row, err := models.CSVUploadTransactions(
		qm.Select("*"),
		qm.OrderBy("timestamp DESC"),
	).One(ctx, h.db.DB)
	if err != nil {
		sugar.Errorf("failed to get csv_upload_transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"timestamp": nil})
		return
	}
	sugar.Debugf("csv_upload_transaction record: %p", row)
	if row == nil {
		sugar.Info("no csv information")
		c.JSON(http.StatusOK, gin.H{"timestamp": nil})
		return
	}

	timestampStr := row.Timestamp.String
	h.log.Info(timestampStr)
	timestampVal := fmt.Sprintf(`%s/%s/%s %s:%s:%s`, timestampStr[0:4], timestampStr[4:6], timestampStr[6:8], timestampStr[8:10], timestampStr[10:12], timestampStr[12:14])
	sugar.Infof("latest timestamp: %s", timestampVal)
	c.JSON(http.StatusOK, gin.H{"timestamp": timestampVal})
	return
}
