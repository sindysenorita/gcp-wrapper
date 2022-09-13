package gcplogger

import (
	"context"
	"fmt"
	"io"
)

func NewStdLog(ctx context.Context, logID string, gcpConfig GCPConfig) (io.Writer, error) {
	gcpLogConfig := GCPLogConfig{
		GCP: GCPConfig{
			ProjectID:          gcpConfig.ProjectID,
			ServiceAccountPath: gcpConfig.ServiceAccountPath,
		},
		LogID: logID,
	}
	// Std log does not use structured logging
	gcpWriter, err := NewWriter(ctx, gcpLogConfig, NoStructuredLogParser)
	if err != nil {
		return nil, fmt.Errorf("failed to create gcplogging writer: %w", err)
	}
	return gcpWriter, nil
}
