package logx

import (
	"github.com/rs/zerolog"
	"os"
	"path/filepath"
	"time"
)

type FileLevelWriter struct{ Dirname, FilePattern string }

var _ zerolog.LevelWriter = (*FileLevelWriter)(nil)

func (w *FileLevelWriter) Write(p []byte) (int, error) { return w.output("", p) }
func (w FileLevelWriter) WriteLevel(l zerolog.Level, p []byte) (int, error) {
	suffix := ""
	switch l {
	case zerolog.WarnLevel, zerolog.ErrorLevel, zerolog.FatalLevel, zerolog.PanicLevel:
		suffix = "-error"
	}
	return w.output(suffix, p)
}
func (w FileLevelWriter) output(suffix string, p []byte) (int, error) {
	if err := os.MkdirAll(w.Dirname, 0750); err != nil {
		return 0, err
	}
	filename := filepath.Join(w.Dirname, time.Now().Format(w.FilePattern)+suffix+".log")
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0640)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	return file.Write(p)
}
