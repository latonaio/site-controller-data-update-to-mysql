package fileController

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
	"site-controller-data-update-to-mysql/app/database"
	"site-controller-data-update-to-mysql/app/file"
	"site-controller-data-update-to-mysql/config"
	"site-controller-data-update-to-mysql/pkg"
)

var sugar = pkg.NewSugaredLogger()

func Watch(ctx context.Context, list chan<- file.Files, done <-chan bool, db *database.Database, env *config.WatchEnv) {
	sugar.Info("created watch go routine")
	// DBから最新のファイルの作成情報を取得する
	rows, err := db.GetCSVUpdateTransaction(ctx)
	if err != nil {
		// log.Errorf("%w", err)
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
			sugar.Infof("received signal: %s", s.String())
			return
		case <-ticker.C:
			sugar.Infof("start watch %s", env.MountPath)
			// ファイルリストの取得
			sugar.Infof("latest file created time: %v", latestFileCreatedTime)
			newFileList, err := file.GetFileList(&latestFileCreatedTime, env.MountPath)
			if err != nil {
				// log.Errorf("cannot get file list in %v: %v", watchDirPath, err)
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
		sugar.Infof("finish watch %s", env.MountPath)
	}
END:
	sugar.Info("finish Watch goroutine")
}
