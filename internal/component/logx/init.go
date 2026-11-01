package logx

import (
	"github.com/labstack/gommon/color"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go-gin/internal/environment"
	"io"
	"strings"
	"time"
)

var (
	AccessLogger zerolog.Logger
	DBLogger     zerolog.Logger
	RestyLogger  zerolog.Logger
	CronLogger   zerolog.Logger
	QueueLogger  zerolog.Logger
)

type Config struct {
	Level string `yaml:"level"`
	Path  string `yaml:"path"`
}

var conf Config

func InitConfig(c Config) {
	if strings.TrimSpace(c.Path) == "" {
		c.Path = "storage/logs"
	}
	conf = c
}
func Init() {
	if environment.IsDebugMode() {
		color.Enable()
	}
	level, err := zerolog.ParseLevel(conf.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.TimeFieldFormat = time.DateTime
	zerolog.LevelFieldMarshalFunc = func(l zerolog.Level) string { return strings.ToUpper(l.String()) }
	log.Logger = newLogger("app", level)
	AccessLogger = newLogger("access", zerolog.InfoLevel)
	// SQL是否记录、错误/慢查询/普通查询由db.log-level独立控制。
	DBLogger = newLogger("sql", zerolog.DebugLevel)
	RestyLogger = newLogger("httpc", zerolog.InfoLevel)
	CronLogger = newLogger("cron", zerolog.InfoLevel)
	QueueLogger = newLogger("queue", zerolog.InfoLevel)
}
func newLogger(category string, level zerolog.Level) zerolog.Logger {
	writer := zerolog.SyncWriter(&moduleWriter{Root: conf.Path, Category: category})
	writers := []io.Writer{writer}
	if environment.IsDebugMode() {
		writers = append(writers, &ConsoleLevelWriter{})
	}
	return zerolog.New(zerolog.MultiLevelWriter(writers...)).Level(level).With().Timestamp().Logger().Hook(TracingHook{})
}
