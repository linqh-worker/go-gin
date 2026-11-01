package logx

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type defaultLog struct {
	ctx    context.Context
	logger *zerolog.Logger
}

var _ Logger = (*defaultLog)(nil)

func WithContext(ctx context.Context) Logger {
	return &defaultLog{ctx: ctx, logger: &log.Logger}
}

func (l *defaultLog) Debug(keyword string, data any) {
	l.print(zerolog.DebugLevel, keyword, data)
}

func (l *defaultLog) Debugf(keyword string, format string, data ...any) {
	l.Debug(keyword, fmt.Sprintf(format, data...))
}

func (l *defaultLog) Info(keyword string, data any) {
	l.print(zerolog.InfoLevel, keyword, data)
}

func (l *defaultLog) Infof(keyword string, format string, data ...any) {
	l.Info(keyword, fmt.Sprintf(format, data...))
}

func (l *defaultLog) Warn(keyword string, data any) {
	l.print(zerolog.WarnLevel, keyword, data)
}

func (l *defaultLog) Warnf(keyword string, format string, data ...any) {
	l.Warn(keyword, fmt.Sprintf(format, data...))
}

func (l *defaultLog) Error(keyword string, data any) {
	l.print(zerolog.ErrorLevel, keyword, data)
}

func (l *defaultLog) Errorf(keyword string, format string, data ...any) {
	l.Error(keyword, fmt.Sprintf(format, data...))
}

func (l *defaultLog) print(level zerolog.Level, keyword string, data any) {
	l.logger.WithLevel(level).Ctx(l.ctx).Str("keyword", keyword).Any("data", data).Send()
}

// WithSQLContext 将数据库日志写入当前模块的sql目录。
func WithSQLContext(ctx context.Context) Logger { return &defaultLog{ctx: ctx, logger: &DBLogger} }
