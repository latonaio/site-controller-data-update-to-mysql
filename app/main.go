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
	"site-controller-data-update-to-mysql/pkg"

	"go.uber.org/zap"
)

func Server(port string, db *database.Database, logger *zap.SugaredLogger) {
	// Server構造体作成
	s := router.NewServer(port, db, logger)
	// Route実行
	s.Route()
	// Server実行
	s.Run()
}

func main() {
	sugar := pkg.NewSugaredLogger()
	defer func(sugar *zap.SugaredLogger) {
		err := sugar.Sync()
		if err != nil {
			fmt.Printf("failed to flush sugar: %v", err)
		}
	}(sugar)

	ctx := context.Background()
	// // Watch内で、新しいファイルが生成されるまで待機させる。
	listAuto := make(chan file.Files)

	// DB構造体作成
	env, err := config.NewEnv()
	if err != nil {
		sugar.Warnf("NewEnv error: %+v", err)
	}
	db, err := database.NewDatabase(env.MysqlEnv)
	if err != nil {
		sugar.Errorf("failed to create database: %+v", err)
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
	go Server(env.Port, db, sugar)

	for {
		select {
		// 自動登録
		case newFileList := <-listAuto:
			for _, file := range newFileList {
				sugar.Infof("target fileName: %v\n", file.Name)

				// csv登録...status＝before
				model, err := db.CreateCsvUploadTransaction(ctx, file.Name, file.CreatedTime, "", "")
				if err != nil {
					sugar.Errorf("failed to insert record to database: %v", err)
				}

				if err := db.RegisterCSVDataToDB(ctx, *file, env.MountPath, model.ID, siteControllerName); err != nil {
					sugar.Error(err)
				}
			}
		case <-quit:
			goto END
		}
	}
END:
	done <- true
	sugar.Info("finish main function")
}
