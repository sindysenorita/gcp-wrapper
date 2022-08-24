package gcp_wrapper

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"io"
)

//type Config struct {
//	Level string
//}

type ZerologGCP struct {
	Logger    zerolog.Logger
	config    GCPLogConfig
	gcpWriter *Writer
}

func NewZerolog(ctx context.Context, logID, level string, gcpConfig GCPConfig) (*ZerologGCP, error) {
	levelMap := map[zerolog.Level]int{
		zerolog.DebugLevel: LevelDebug,
		zerolog.InfoLevel:  LevelInfo,
		zerolog.WarnLevel:  LevelWarning,
		zerolog.ErrorLevel: LevelError,
		zerolog.PanicLevel: LevelCritical,
		zerolog.FatalLevel: LevelCritical,
	}
	levelMod := LevelModifier{
		OriginalField:  zerolog.LevelFieldName,
		RemoveOriginal: true,
		Mapping: func(originalLvl interface{}) int {
			ori, ok := originalLvl.(string)
			if !ok {
				return int(LevelDefault)
			}
			zlogLvl, _ := zerolog.ParseLevel(ori)
			return levelMap[zlogLvl]
		},
	}
	gcpLogConfig := GCPLogConfig{
		GCP: GCPConfig{
			ProjectID:          gcpConfig.ProjectID,
			ServiceAccountPath: gcpConfig.ServiceAccountPath,
		},
		LogID:         logID,
		LevelModifier: levelMod,
		Level:         level,
	}
	gcpWriter, err := NewWriter(ctx, GCPLogConfig{
		GCP: GCPConfig{
			ProjectID:          gcpConfig.ProjectID,
			ServiceAccountPath: gcpConfig.ServiceAccountPath,
		},
		LogID:         logID,
		LevelModifier: levelMod,
	})
	z := &ZerologGCP{config: gcpLogConfig}
	if err != nil {
		return nil, fmt.Errorf("failed to create gcplogging writer: %w", err)
	}
	lg, err := newZerolog(gcpLogConfig, gcpWriter)
	if err != nil {
		return nil, err
	}
	z.gcpWriter = gcpWriter
	z.Logger = lg
	return z, nil
}

func newZerolog(cfg GCPLogConfig, writer io.Writer) (zerolog.Logger, error) {
	logLvl, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		return zerolog.Nop(), fmt.Errorf("invalid zerolog log level: %w", err)
	}
	zlog := zerolog.New(writer).
		Level(logLvl).
		With().
		Str("service", cfg.LogID).
		Timestamp().
		Logger()
	return zlog, nil
}

func (z *ZerologGCP) Flush() {
	if z.gcpWriter != nil {
		z.gcpWriter.Flush()
	}
}
