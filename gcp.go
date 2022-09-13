package gcplogger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"cloud.google.com/go/logging"
	"google.golang.org/api/option"
)

type GCPLogConfig struct {
	GCP           GCPConfig
	LogID         string
	LevelModifier LevelModifier
}
type GCPConfig struct {
	ProjectID          string
	ServiceAccountPath string
}

// Specify to modify logfields before sending to GCP logging.
// The most common case is to map  other logging's library level to GCP's, read: https://github.com/rs/zerolog/issues/174
//
// OriginalField: field key of other logging's log level
// RemoveOriginal: remove original level field or not
// Mapping: function to convert other logging's log level value to GCP's log level
//
// list of GCP log level in: https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#logseverity
type LevelModifier struct {
	OriginalField  string
	RemoveOriginal bool
	Mapping        func(originalLvl interface{}) logging.Severity
}

type StructuredLog map[string]interface{}
type StructuredLogParser func(b []byte) (StructuredLog, error)

type Writer struct {
	cfg        GCPLogConfig
	logger     *logging.Logger
	sLogParser StructuredLogParser
}

// TODO use better auth method than service account
func NewWriter(
	ctx context.Context,
	cfg GCPLogConfig,
	sLogParser StructuredLogParser,
) (*Writer, error) {
	client, err := logging.NewClient(
		ctx,
		cfg.GCP.ProjectID,
		option.WithCredentialsFile(cfg.GCP.ServiceAccountPath),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to init stackdriver NewClient: %w", err)
	}
	s := &Writer{
		cfg:    cfg,
		logger: client.Logger(cfg.LogID),
	}
	if sLogParser != nil {
		s.sLogParser = sLogParser
	} else {
		s.sLogParser = JsonStructuredLogParser
	}
	return s, nil
}

func NoStructuredLogParser(b []byte) (StructuredLog, error) {
	return StructuredLog{
		"message": string(b),
	}, nil
}

func JsonStructuredLogParser(b []byte) (StructuredLog, error) {
	var logFields StructuredLog
	err := json.NewDecoder(bytes.NewReader(b)).Decode(&logFields)
	if err != nil {
		return nil, fmt.Errorf("failed to decode logFields: %w", err)
	}
	return logFields, nil
}

func (s *Writer) Write(b []byte) (n int, err error) {
	entry := logging.Entry{}
	if s.sLogParser == nil {
		return 0, fmt.Errorf("nil structured log parser")
	}
	logFields, err := s.sLogParser(b)
	if err != nil {
		return 0, fmt.Errorf("error from structured log parser: %w", err)
	}

	mod := s.cfg.LevelModifier
	// if true, need to modify the severity field in the original data
	if mod.Mapping != nil || mod.RemoveOriginal {
		oriLvl, ok := logFields[mod.OriginalField]
		if ok {
			entry.Severity = mod.Mapping(oriLvl)
		}
	}
	if s.cfg.LevelModifier.RemoveOriginal {
		delete(logFields, mod.OriginalField)
	}
	entry.Payload = logFields
	s.logger.Log(entry)
	return len(b), nil
}
