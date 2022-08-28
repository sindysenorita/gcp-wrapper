package gcplogger

import (
	"cloud.google.com/go/logging"
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"io"
)

type ZerologGCP struct {
	Logger    zerolog.Logger
	config    GCPLogConfig
	gcpWriter *Writer
}

func NewZerolog(ctx context.Context, logID string, gcpConfig GCPConfig) (io.Writer, error) {
	levelMap := map[string]logging.Severity{
		zerolog.DebugLevel.String(): logging.Debug,
		zerolog.InfoLevel.String():  logging.Info,
		zerolog.WarnLevel.String():  logging.Warning,
		zerolog.ErrorLevel.String(): logging.Error,
		zerolog.PanicLevel.String(): logging.Critical,
		zerolog.FatalLevel.String(): logging.Critical,
		zerolog.NoLevel.String():    logging.Default,
	}
	levelMod := LevelModifier{
		OriginalField:  zerolog.LevelFieldName,
		RemoveOriginal: true,
		Mapping: func(originalLvl interface{}) logging.Severity {
			oriString, ok := originalLvl.(string)
			if !ok {
				return logging.Default
			}
			gcpSeverity, ok := levelMap[oriString]
			if !ok {
				return logging.Default
			}
			return gcpSeverity
		},
	}
	gcpLogConfig := GCPLogConfig{
		GCP: GCPConfig{
			ProjectID:          gcpConfig.ProjectID,
			ServiceAccountPath: gcpConfig.ServiceAccountPath,
		},
		LogID:         logID,
		LevelModifier: levelMod,
	}
	gcpWriter, err := NewWriter(ctx, gcpLogConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create gcplogging writer: %w", err)
	}

	return gcpWriter, nil
}
