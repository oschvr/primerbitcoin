package utils

import (
	"github.com/sirupsen/logrus"
	"os"
)

var Logger = logrus.New()

func init() {

	// Output logs to stdout
	Logger.SetOutput(os.Stdout)

	// Set debug level
	Logger.SetLevel(logrus.DebugLevel)

	// Use the full timestamp
	Logger.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	})
}
