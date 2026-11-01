package logx

import (
	"context"
	"encoding/json"
	"github.com/rs/zerolog"
	"path/filepath"
	"sync/atomic"
	"time"
)

type moduleKey struct{}

var defaultModule atomic.Value

func validModule(module string) bool {
	return module == "api" || module == "system" || module == "common"
}

// SetDefaultModule 设置进程缺省归属；all及无归属任务写common。请求不修改此值。
func SetDefaultModule(module string) {
	if !validModule(module) {
		module = "common"
	}
	defaultModule.Store(module)
}
func DefaultModule() string {
	if module, ok := defaultModule.Load().(string); ok {
		return module
	}
	return "common"
}

// WithModule 使用受控枚举，禁止将任意请求值拼成目录。
func WithModule(ctx context.Context, module string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if !validModule(module) {
		module = DefaultModule()
	}
	return context.WithValue(ctx, moduleKey{}, module)
}
func Module(ctx context.Context) string {
	if ctx != nil {
		if module, ok := ctx.Value(moduleKey{}).(string); ok && validModule(module) {
			return module
		}
	}
	return DefaultModule()
}

// moduleWriter 从日志钩子生成的模块字段选择目录，兼容已有各类zerolog日志器。
type moduleWriter struct{ Root, Category string }

func (w *moduleWriter) Write(p []byte) (int, error) { return w.WriteLevel(zerolog.NoLevel, p) }
func (w *moduleWriter) WriteLevel(level zerolog.Level, p []byte) (int, error) {
	var record struct {
		Module string `json:"module"`
	}
	if err := json.Unmarshal(p, &record); err != nil {
		return 0, err
	}
	module := record.Module
	if !validModule(module) {
		module = "common"
	}
	writer := FileLevelWriter{Dirname: filepath.Join(w.Root, module, w.Category) + string(filepath.Separator), FilePattern: time.DateOnly}
	return writer.WriteLevel(level, p)
}
