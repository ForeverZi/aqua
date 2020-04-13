package log

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Logger *zap.Logger
var Sugar *zap.SugaredLogger
var Level zap.AtomicLevel
var Hook lumberjack.Logger

func init() {
	Hook = lumberjack.Logger{
		Filename:   "./logfiles/aqua.log", // 日志文件路径
		MaxSize:    128,              // 每个日志文件保存的最大尺寸 单位：M
		MaxBackups: 30,               // 日志文件最多保存多少个备份
		MaxAge:     7,                // 文件最多保存多少天
		Compress:   false,            // 是否压缩
	}
	Level = zap.NewAtomicLevel()
	fileCore := zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(&Hook), Level)
	consoleCore := zapcore.NewCore(zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()), zapcore.Lock(zapcore.AddSync(os.Stdout)), Level)
	Logger = zap.New(zapcore.NewTee(fileCore, consoleCore), zap.AddCaller(), zap.AddCallerSkip(1), zap.Development())
	Sugar = Logger.Sugar()
}

func New() *ZapSugarLogger{
	return &ZapSugarLogger{
		s: Sugar,
	}
}

type ZapSugarLogger struct{
	s   	*zap.SugaredLogger
}

func (l *ZapSugarLogger) Println(msg string, v ...interface{}){
	l.s.Infow(msg, v...)
}

func Clean() {
	Logger.Sync()
	Hook.Close()
}