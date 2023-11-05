package audit

import (
	"log/slog"
	"os"
)

type Logger struct {
	underlying *slog.Logger
	enabled    bool
	file       *os.File
}

func New(enabled bool) (*Logger, error) {
	if !enabled {
		return &Logger{enabled: false}, nil
	}

	file, err := os.OpenFile(os.TempDir()+"/audit.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &Logger{
		enabled:    true,
		underlying: slog.New(slog.NewJSONHandler(file, &slog.HandlerOptions{Level: slog.LevelDebug})),
		file:       file,
	}, nil
}

func (l *Logger) Log(id uint16, type_ uint16, query []byte, response []byte) {
	if !l.enabled {
		return
	}

	l.underlying.Debug("AUDIT", "id", id, "type", type_, "query", query, "response", response)
}

func (l *Logger) Close() error {
	if !l.enabled {
		return nil
	}

	return l.file.Close()
}
