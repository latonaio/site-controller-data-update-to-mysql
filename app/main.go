package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"site-controller-data-update-to-mysql/app/cmd/fileController"
	"site-controller-data-update-to-mysql/app/database"
	"site-controller-data-update-to-mysql/app/file"
	"site-controller-data-update-to-mysql/app/server/router"
	"site-controller-data-update-to-mysql/config"

	"github.com/latonaio/golang-logging-library/logger"
)

func Server(port string, db *database.Database, logging *logger.Logger) {
	// Server構造体作成
	s := router.NewServer(port, db, logging)
	// Route実行
	s.Route()
	// Server実行
	s.Run()
}

func main() {
	var logging = logger.NewLogger()
	ctx := context.Background()
	// // Watch内で、新しいファイルが生成されるまで待機させる。
	listAuto := make(chan file.Files)

	// DB構造体作成
	env, err := config.NewEnv()
	if err != nil {
		logging.Warn(fmt.Sprintf("NewEnv error: %+v", err), nil)
	}
	db, err := database.NewDatabase(env.MysqlEnv)
	if err != nil {
		logging.Error(fmt.Sprintf("failed to create database: %+v", err), nil)
		return
	}
	// mainを終了させるためのチャネル
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	// mainが終了した時にgoルーチンも終了するためのチャネル
	done := make(chan bool, 1)

	siteControllerName := config.GetEnv("SITE_CONTROLLER_NAME", "Lincoln")

	// 自動でcsvファイルからデータをMySQLに入れるgoルーチン
	go fileController.Watch(ctx, listAuto, done, db, env.WatchEnv)

	// HTTPサーバを立てる
	go Server(env.Port, db, logging)

	for {
		select {
		// 自動登録
		case newFileList := <-listAuto:
			for _, file := range newFileList {
				logging.Info(fmt.Sprintf("target fileName: %v\n", file.Name), nil)

				// csv登録...status＝before
				model, err := db.CreateCsvUploadTransaction(ctx, file.Name, file.CreatedTime, "", "")
				if err != nil {
					logging.Error(fmt.Sprintf("failed to insert record to database: %v", err), nil)
				}

				if err := db.RegisterCSVDataToDB(ctx, *file, env.MountPath, model.ID, siteControllerName); err != nil {
					logging.Error(err, nil)
				}
			}
		case <-quit:
			goto END
		}
	}
END:
	done <- true
	logging.Info("finish main function", nil)

}
