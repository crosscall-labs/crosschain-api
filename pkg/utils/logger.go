package utils

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

func FormatKeyValueLogs(data [][2]string) string {
	var builder strings.Builder
	builder.Grow(len(data) * 10)

	for _, entry := range data {
		builder.WriteString(fmt.Sprintf("  %s: %s\n", entry[0], entry[1]))
	}

	return builder.String()
}

func LogInfo(title string, message string) {
	logrus.Info(fmt.Sprintf(
		"\033[1m%s\033[0m:\n%s",
		title,
		message,
	))
}

func LogInfoSimple(message string) {
	logrus.Info(message)
}

func LogError(message string, errStr string) {
	logrus.Error(fmt.Sprintf(
		"%s: \033[38;5;197m%s",
		message,
		errStr,
	))
}

func LogNotice(message string) {
	logrus.Warn(message)
}
