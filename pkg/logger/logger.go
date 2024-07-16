package logger

import (
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"os"
	"strings"
	"time"
	"uc/pkg/nacos"
)

var Logger *zap.SugaredLogger

func Init() {

	encoderConfig := zapcore.EncoderConfig{
		MessageKey:       "msg",
		LevelKey:         "level",
		TimeKey:          "ts",
		NameKey:          "log",
		CallerKey:        "caller",
		FunctionKey:      "",
		StacktraceKey:    "stacktrace",
		SkipLineEnding:   false,
		LineEnding:       "\n",
		EncodeLevel:      zapcore.CapitalLevelEncoder,
		EncodeTime:       zapcore.ISO8601TimeEncoder,
		EncodeDuration:   zapcore.SecondsDurationEncoder,
		EncodeCaller:     zapcore.ShortCallerEncoder,
		EncodeName:       zapcore.FullNameEncoder,
		ConsoleSeparator: " ",
	}
	jsonEncoder := zapcore.NewJSONEncoder(encoderConfig)

	infoLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl == zapcore.InfoLevel
	})
	errorLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.WarnLevel
	})
	//
	errorWriter := getWriter(nacos.Config.Log.ErrorPath)
	infoWriter := getWriter(nacos.Config.Log.InfoPath)

	//infoWriter := getWriter("info")
	//errorWriter := getWriter("error")
	core := zapcore.NewTee(
		zapcore.NewCore(jsonEncoder, zapcore.AddSync(infoWriter), infoLevel),
		zapcore.NewCore(jsonEncoder, zapcore.AddSync(errorWriter), errorLevel),
	)
	log := zap.New(core, zap.Fields(zap.String("product_name", "uc")), zap.AddCaller())
	Logger = log.Sugar().WithOptions(zap.AddCallerSkip(1))
}

func getWriter(filename string) io.Writer {
	hook, err := rotatelogs.New(
		// demo.YYmmddHH.log
		strings.Replace(filename, ".log", "", -1)+"-%Y%m%d%H.log",
		// 保存xxx小时
		rotatelogs.WithMaxAge(time.Hour*time.Duration(nacos.Config.Log.MaxAge)),
		// 按xxx小时切割
		rotatelogs.WithRotationTime(time.Hour*time.Duration(nacos.Config.Log.Rotation)),
	)
	if err != nil {
		os.Exit(202)
	}
	return hook
}

func Debug(args ...interface{}) {
	Logger.Debug(args...)
}

func Debugf(template string, args ...interface{}) {
	Logger.Debugf(template, args...)
}

func Info(args ...interface{}) {
	Logger.Info(args...)
}

func Infof(template string, args ...interface{}) {
	Logger.Infof(template, args...)
}

func Warn(args ...interface{}) {
	Logger.Warn(args...)
}

func Warnf(template string, args ...interface{}) {
	Logger.Warnf(template, args...)
}

func Error(args ...interface{}) {
	Logger.Error(args...)
}

func Errorf(template string, args ...interface{}) {
	Logger.Errorf(template, args...)
}

func DPanic(args ...interface{}) {
	Logger.DPanic(args...)
}

func DPanicf(template string, args ...interface{}) {
	Logger.DPanicf(template, args...)
}

func Panic(args ...interface{}) {
	Logger.Panic(args...)
}

func Panicf(template string, args ...interface{}) {
	Logger.Panicf(template, args...)
}

func Fatal(args ...interface{}) {
	Logger.Fatal(args...)
}

func Fatalf(template string, args ...interface{}) {
	Logger.Fatalf(template, args...)
}
