// SPDX-FileCopyrightText: 2025 Jacques Supcik <jacques.supcik@hefr.ch>
//
// SPDX-License-Identifier: MIT

package cmd

import (
	"context"
	"fmt"
	"log/slog"
)

type MqttLogger struct {
	level  slog.Level
	logger *slog.Logger
}

func NewMqttLogger(level slog.Level, logger *slog.Logger) *MqttLogger {
	return &MqttLogger{
		level:  level,
		logger: logger,
	}
}

func (l *MqttLogger) Printf(format string, v ...interface{}) {
	l.logger.Log(context.Background(), l.level, fmt.Sprintf(format, v...))
}

func (l *MqttLogger) Println(v ...interface{}) {
	l.logger.Log(context.Background(), l.level, fmt.Sprint(v...))
}
