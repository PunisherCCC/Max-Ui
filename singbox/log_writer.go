package singbox

import (
	"regexp"
	"strings"

	"github.com/mhsanaei/3x-ui/v2/logger"
)

type LogWriter struct {
	lastLine string
}

func NewLogWriter() *LogWriter {
	return &LogWriter{}
}

func (lw *LogWriter) LastLine() string {
	if lw == nil {
		return ""
	}
	return lw.lastLine
}

func (lw *LogWriter) Write(m []byte) (int, error) {
	crashRegex := regexp.MustCompile(`(?i)(panic|exception|stack trace|fatal error)`)
	message := strings.TrimSpace(string(m))
	if crashRegex.MatchString(message) {
		logger.Debug("Sing-box crash detected:\n", message)
		lw.lastLine = message
		return len(m), nil
	}
	for _, msg := range strings.Split(message, "\n") {
		msg = strings.TrimSpace(msg)
		if msg == "" {
			continue
		}
		if strings.Contains(strings.ToLower(msg), "error") || strings.Contains(strings.ToLower(msg), "failed") {
			logger.Error("SING-BOX: " + msg)
		} else {
			logger.Debug("SING-BOX: " + msg)
		}
		lw.lastLine = msg
	}
	return len(m), nil
}
