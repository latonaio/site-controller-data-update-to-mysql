package router

import (
	"fmt"
	"log"
	"site-controller-data-update-to-mysql/app/database"
	"site-controller-data-update-to-mysql/app/server/handlers"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/latonaio/golang-logging-library/logger"
)

type Server struct {
	gin  *gin.Engine
	ws   *websocket.Upgrader
	port string
	db   *database.Database
	log  *logger.Logger
}

func NewServer(port string, db *database.Database, logging *logger.Logger) *Server {
	return &Server{
		gin: gin.New(),
		ws: &websocket.Upgrader{
			HandshakeTimeout:  5 * time.Second,
			ReadBufferSize:    0,
			WriteBufferSize:   0,
			WriteBufferPool:   nil,
			Subprotocols:      nil,
			Error:             nil,
			CheckOrigin:       nil,
			EnableCompression: false,
		},
		port: fmt.Sprintf(`:%v`, port),
		db:   db,
		log:  logging,
	}
}

func (s *Server) Route() {
	handler := handlers.NewSCHandler(s.db, s.log)

	s.gin.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"*",
		},
		AllowMethods: []string{
			"POST",
			"GET",
			"PUT",
			"OPTIONS",
			"DELETE",
		},
		AllowHeaders: []string{
			"Accept",
			"Authorization",
			"Content-type",
			"X-CSRF-Token",
		},
		ExposeHeaders: []string{
			"Link",
		},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	// restAPI一覧
	s.gin.GET("/api/auth/csv/:timestamp", handler.GetAuthCSV)

	baseGroup := s.gin.Group("/api/csv")

	// エラーハンドリング用エンドポイントその１
	baseGroup.GET("/transaction/display/errors", handler.CSVError)

	// エラーテーブルにあるエラーを全て解決済みに更新する
	baseGroup.POST("/error/resolve", handler.UpdateErrorStatus)

	// csv手動登録用エンドポイント
	baseGroup.POST("/:timestamp", handler.CreateCSV)

	// 前回の手動連携日時を返すエンドポイント
	baseGroup.GET("/transaction/latest", handler.GetLatestTimestamp)

	//g.GET("/:timestamp")
	//g.POST("/:timestamp")

	// TODO エラーハンドリング用エンドポイントその２
	s.gin.GET("/ws", func(c *gin.Context) {
		// handlers.GetCSVErrorHandler(c.Writer, c.Request)
		// handlers.WsConnect(c, s.db)
		// handlers.WsConnect(c, s.db, channel)
	})
}

func (s *Server) Run() {
	// log.Println("run server")
	if err := s.gin.Run(s.port); err != nil {
		log.Printf("run server error :%v", err)
	}
}
