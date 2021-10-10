package pkg

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewSugaredLogger() *zap.SugaredLogger {
	// ログレベル
	level := zap.NewAtomicLevel()
	level.SetLevel(zapcore.InfoLevel)

	// コンフィグ
	myConfig := zap.Config{
		Level:             level,
		Development:       false,
		DisableCaller:     true,
		DisableStacktrace: true,
		Sampling:          nil,
		Encoding:          "console",
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey: "Msg",
			LevelKey:   "Level",
			TimeKey:    "Time",
			NameKey:    "Name",
			CallerKey:  "Caller",
			//FunctionKey:      "",
			StacktraceKey: "St",
			//LineEnding:       "",
			EncodeLevel:    zapcore.CapitalColorLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
			//EncodeName:       nil,
			//ConsoleSeparator: "",
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		//InitialFields:     nil,
	}

	logger, _ := myConfig.Build()
	return logger.Sugar()
}
