package handlers

import (
	"encoding/json"
	"log"
	"site-controller-data-update-to-mysql/app/server/response"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (h *SCHandler) WsConnect(c *gin.Context, channel chan []int) {
	conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to set websocket upgrade: %v\n", err)
		return
	}
	tx, err := h.db.DB.Begin()
	//何か受け取ってそのまま返すパターン
END:
	for {
		select {
		case errorRowIds := <-channel:
			rows, err := h.db.SelectErrorCSVRowsWithIds(errorRowIds, c.Request.Context(), tx)
			if err != nil {
				logging.Error("cannot get csv error rows", nil)
				break END
			}
			row := rows[0]
			responseStruct := response.CsvExectutionError{
				FileName: row.R.CSV.FileName.String,
				Errors: []response.Error{
					{
						LineNumber:          row.LineNumber,
						CustomerName:        row.CustomerName.String,
						CustomerPhoneNumber: row.CustomerPhoneNumber.String,
					},
				},
			}

			res, err := json.Marshal(responseStruct)
			if err != nil {
				logging.Error(err, nil)
				break END
			}
			conn.WriteMessage(websocket.BinaryMessage, res)
		}
	}
}

func sendError() {

}

func sendNoneError() {

}
