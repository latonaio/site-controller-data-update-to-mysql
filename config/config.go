package config

import (
	"fmt"
	"os"
	"strconv"

	"golang.org/x/xerrors"
)

type Env struct {
	*MysqlEnv
	*WatchEnv
	Port string
}
type MysqlEnv struct {
	User     xxxx
	Host     xxxx
	Password xxxx
	Port     xxxx
}

type WatchEnv struct {
	PollingInterval int
	MountPath       string
}

// NewEnv 必ずEnv構造体は返る、POLLING_INTERVALに数字が入っていない場合にエラーが返る
func NewEnv() (*Env, error) {
	watchEnv, err := NewWatchEnv()
	return &Env{
		MysqlEnv: NewMysqlEnv(),
		WatchEnv: watchEnv,
		Port:     GetEnv("PORT", "8080"),
	}, err
}

func NewMysqlEnv() *MysqlEnv {
	user := GetEnv("MYSQL_USER", "xxxx")
	host := GetEnv("MYSQL_HOST", "xxxx")
	pass := GetEnv("MYSQL_PASSWORD", "xxxx")
	port := GetEnv("MYSQL_PORT", "xxxx")

	return &MysqlEnv{
		User:     user,
		Host:     host,
		Password: pass,
		Port:     port,
	}
}

func NewWatchEnv() (*WatchEnv, error) {
	pollingInterval, err := strconv.Atoi(GetEnv("POLLING_INTERVAL", "1"))
	if err != nil {
		pollingInterval = 1
		err = xerrors.Errorf("POLLING_INTERVAL should be int: %w", err)
	}
	return &WatchEnv{
		PollingInterval: pollingInterval,
		MountPath:       GetEnv("MOUNT_PATH", "/mnt/windows"),
	}, err
}

func (c *MysqlEnv) DSN() string {
	return fmt.Sprintf(`%v:%v@tcp(%v:%v)/%s?charset=utf8mb4&parseTime=True&loc=Local`, c.User, c.Password, c.Host, c.Port, "xxxx")
}

func GetEnv(key, def string) string {
	value := os.Getenv(key)
	if value == "" {
		value = def
	}
	return value
}
