package utils

import (
	"os"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

var Logger = log.New(os.Stderr)

func PrepareLogger() {
	loggerStyles := log.DefaultStyles()
	loggerStyles.Levels[log.FatalLevel] = lipgloss.NewStyle().SetString("FATL").Bold(true).Foreground(lipgloss.Color("196"))
	loggerStyles.Levels[log.ErrorLevel] = lipgloss.NewStyle().SetString("ERRO").Bold(true).Foreground(lipgloss.Color("202"))
	loggerStyles.Levels[log.WarnLevel] = lipgloss.NewStyle().SetString("WARN").Bold(true).Foreground(lipgloss.Color("214"))
	loggerStyles.Levels[log.InfoLevel] = lipgloss.NewStyle().SetString("INFO").Bold(true).Foreground(lipgloss.Color("40"))
	loggerStyles.Levels[log.DebugLevel] = lipgloss.NewStyle().SetString("DEBG").Bold(true).Foreground(lipgloss.Color("33"))
	Logger.SetReportTimestamp(true)
	Logger.SetReportCaller(true)
	Logger.SetTimeFormat(time.DateTime)
	Logger.SetLevel(log.DebugLevel)
	Logger.SetStyles(loggerStyles)
}
