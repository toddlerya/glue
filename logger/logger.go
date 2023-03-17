package logger

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	log "github.com/sirupsen/logrus"
)

func InitLogConfig(formatter, logPath, logName string) {
	if formatter == "json" {
		//设置为json格式
		log.SetFormatter(
			&log.JSONFormatter{
				TimestampFormat: "2006-01-02 15:04:05",
			})
	} else if formatter == "text" {
		// 文本格式
		log.SetFormatter(&log.TextFormatter{
			FullTimestamp:          true,
			TimestampFormat:        "2006-01-02 15:04:05",
			ForceColors:            true,
			DisableLevelTruncation: true,
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				funcName := filepath.Base(f.Function)
				return funcName + "()", filepath.Base(f.File) + ":" + strconv.Itoa(f.Line)
			},
		})
	} else {
		log.Panic("日志格式配置错误: ", formatter)
	}

	fileFormatter := &log.TextFormatter{
		FullTimestamp:          true,
		TimestampFormat:        "2006-01-02 15:04:05",
		ForceColors:            false,
		DisableLevelTruncation: true,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			funcName := filepath.Base(f.Function)
			return funcName + "()", f.File + ":" + strconv.Itoa(f.Line)
		},
	}

	logFilePath := filepath.Join(logPath, logName)
	writer, err := rotatelogs.New(
		logFilePath+".%Y%m%d%H",
		// WithLinkName为最新的日志建立软连接,以方便随着找到当前日志文件
		rotatelogs.WithLinkName(logFilePath),
		// WithRotationTime设置日志分割的时间,这里设置为24*7小时分割一次
		rotatelogs.WithRotationTime(time.Hour*24*7),
		// WithMaxAge和WithRotationCount二者只能设置一个,
		// WithMaxAge设置文件清理前的最长保存时间,
		// WithRotationCount设置文件清理前最多保存的个数.
		// rotatelogs.WithMaxAge(time.Hour*24),
		rotatelogs.WithRotationCount(10),
	)
	if err != nil {
		log.Errorf("config local file system for logger error: %v", err)
		panic(err)
	}

	lfHook := lfshook.NewHook(lfshook.WriterMap{
		log.DebugLevel: writer,
		log.InfoLevel:  writer,
		log.WarnLevel:  writer,
		log.ErrorLevel: writer,
		log.FatalLevel: writer,
		log.PanicLevel: writer,
	}, fileFormatter)
	log.AddHook(lfHook)

	// 设置将日志输出到标准输出（默认的输出为stderr,标准错误）
	// 日志消息输出可以是任意的io.writer类型
	mw := io.MultiWriter(os.Stdout)
	log.SetOutput(mw)

}

func SetLogLevel(level string) {
	level = strings.ToUpper(level)
	log.Infof("Set log level %s", level)
	switch level {
	case "TRACE":
		log.SetLevel(log.TraceLevel)
		log.SetReportCaller(true)
	case "DEBUG":
		log.SetLevel(log.DebugLevel)
		log.SetReportCaller(true)
	case "INFO":
		log.SetLevel(log.InfoLevel)
	case "WARN":
		log.SetLevel(log.WarnLevel)
	case "ERROR":
		log.SetLevel(log.ErrorLevel)
	case "FATAL":
		log.SetLevel(log.FatalLevel)
	case "PANIC":
		log.SetLevel(log.PanicLevel)
	}
}
