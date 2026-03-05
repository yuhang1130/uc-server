package logger

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/DeRuina/timberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Logger() *zap.Logger

	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})

	// 格式化日志方法（SugaredLogger 支持）
	Debugf(template string, args ...interface{})
	Infof(template string, args ...interface{})
	Warnf(template string, args ...interface{})
	Errorf(template string, args ...interface{})
	Fatalf(template string, args ...interface{})

	// 键值对结构化日志（SugaredLogger 支持）
	Debugw(msg string, keysAndValues ...interface{})
	Infow(msg string, keysAndValues ...interface{})
	Warnw(msg string, keysAndValues ...interface{})
	Errorw(msg string, keysAndValues ...interface{})
	Fatalw(msg string, keysAndValues ...interface{})
}

type loggerImpl struct {
	logger *zap.Logger
	sugar  *zap.SugaredLogger
}

// 创建一个类型为 *loggerImpl 的 nil 指针
// 将它赋值给 Logger 接口类型，_ 表示丢弃这个变量（不实际使用）
var _ Logger = (*loggerImpl)(nil)

func New(dir string, level string, isProd bool) Logger {
	var l zapcore.Level

	switch strings.ToLower(level) {
	case "error":
		l = zap.ErrorLevel
	case "warn":
		l = zap.WarnLevel
	case "info":
		l = zap.InfoLevel
	case "debug":
		l = zap.DebugLevel
	default:
		l = zap.InfoLevel
	}

	var consoleCfg zapcore.EncoderConfig
	if isProd {
		consoleCfg = zap.NewProductionEncoderConfig()
		consoleCfg.EncodeLevel = zapcore.CapitalLevelEncoder
	} else {
		consoleCfg = zap.NewDevelopmentEncoderConfig()
		consoleCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	consoleCfg.EncodeTime = zapcore.RFC3339TimeEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(consoleCfg)

	cores := []zapcore.Core{
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), l),
	}

	if dir != "" {
		fileCfg := zap.NewProductionEncoderConfig()
		fileCfg.EncodeTime = zapcore.RFC3339TimeEncoder
		fileEncoder := zapcore.NewJSONEncoder(fileCfg)

		appCore := zapcore.NewCore(fileEncoder, zapcore.AddSync(getLogWriter(filepath.Join(dir, "app.log"))), l)
		errCore := zapcore.NewCore(fileEncoder, zapcore.AddSync(getLogWriter(filepath.Join(dir, "app-error.log"))), zap.ErrorLevel)
		cores = append(cores, appCore, errCore)
	}

	base := zap.New(zapcore.NewTee(cores...), zap.AddCaller(), zap.AddCallerSkip(1))
	return &loggerImpl{logger: base, sugar: base.Sugar()}
}

func getLogWriter(filename string) io.Writer {
	return &timberjack.Logger{
		Filename:           filename,
		MaxSize:            200,
		MaxBackups:         3,
		MaxAge:             14,
		Compression:        "gzip",
		LocalTime:          true,
		RotationInterval:   24 * time.Hour,
		RotateAt:           []string{"00:00", "12:00"},
		BackupTimeFormat:   "2006-01-02-15-04-05",
		AppendTimeAfterExt: true,
	}
}

func (l *loggerImpl) Logger() *zap.Logger { return l.logger }

func (l *loggerImpl) Debug(args ...interface{}) { l.sugar.Debug(args...) }
func (l *loggerImpl) Info(args ...interface{})  { l.sugar.Info(args...) }
func (l *loggerImpl) Warn(args ...interface{})  { l.sugar.Warn(args...) }
func (l *loggerImpl) Error(args ...interface{}) { l.sugar.Error(args...) }
func (l *loggerImpl) Fatal(args ...interface{}) { l.sugar.Fatal(args...) }

// 格式化日志方法
func (l *loggerImpl) Debugf(t string, a ...interface{}) { l.sugar.Debugf(t, a...) }
func (l *loggerImpl) Infof(t string, a ...interface{})  { l.sugar.Infof(t, a...) }
func (l *loggerImpl) Warnf(t string, a ...interface{})  { l.sugar.Warnf(t, a...) }
func (l *loggerImpl) Errorf(t string, a ...interface{}) { l.sugar.Errorf(t, a...) }
func (l *loggerImpl) Fatalf(t string, a ...interface{}) { l.sugar.Fatalf(t, a...) }

// 键值对结构化日志方法
func (l *loggerImpl) Debugw(msg string, kv ...interface{}) { l.sugar.Debugw(msg, kv...) }
func (l *loggerImpl) Infow(msg string, kv ...interface{})  { l.sugar.Infow(msg, kv...) }
func (l *loggerImpl) Warnw(msg string, kv ...interface{})  { l.sugar.Warnw(msg, kv...) }
func (l *loggerImpl) Errorw(msg string, kv ...interface{}) { l.sugar.Errorw(msg, kv...) }
func (l *loggerImpl) Fatalw(msg string, kv ...interface{}) { l.sugar.Fatalw(msg, kv...) }
