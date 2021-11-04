package fileController

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"site-controller-data-update-to-mysql/app/database"
	"site-controller-data-update-to-mysql/app/file"
	"site-controller-data-update-to-mysql/config"
	"syscall"
	"time"

	"github.com/latonaio/golang-logging-library/logger"
)

var logging = logger.NewLogger()

func Watch(ctx context.Context, list chan<- file.Files, done <-chan bool, db *database.Database, env *config.WatchEnv) {
	logging.Info("created watch go routine", nil)
	// DBから最新のファイルの作成情報を取得する
	rows, err := db.GetCSVUpdateTransaction(ctx)
	if err != nil {
	}
	var latestFileCreatedTime time.Time
	if len(rows) > 0 {
		createdTimeInWindows := rows[0].CreatedTimeInWindows
		if t, _ := createdTimeInWindows.Value(); t != nil {
			latestFileCreatedTime = t.(time.Time)
		}
	}

	tickTime := time.Duration(env.PollingInterval) * time.Minute
	ticker := time.NewTicker(tickTime)
	defer ticker.Stop()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGTERM)

	for {
		select {
		case s := <-signalCh:
			logging.Info(fmt.Sprintf("received signal: %s", s.String()), nil)

			return
		case <-ticker.C:
			logging.Info(fmt.Sprintf("start watch %s ", env.MountPath), nil)
			logging.Info(fmt.Sprintf("latest file created time: %v", latestFileCreatedTime), nil)

			newFileList, err := file.GetFileList(&latestFileCreatedTime, env.MountPath)
			if err != nil {
				goto L
			}

			if len(newFileList) == 0 {
				goto L
			}

			// ファイル登録処理へ渡す
			list <- newFileList

			// 最新ファイルの更新
			latestFileCreatedTime = newFileList[0].CreatedTime
		case <-done:
			goto END
		}
	L:
		logging.Info(fmt.Sprintf("finish watch %s", env.MountPath), nil)

	}
END:
	logging.Info("finish Watch goroutine", nil)
}
