package log_test

import (
	l "log"

	"github.com/adamluo159/leaf/log"
)

func Example() {
	name := "Leaf"

	log.Debug("My name is %v", name)
	log.Release("My name is %v", name)
	log.Error("My name is %v", name)
	// log.Fatal("My name is %v", name)

	logger, err := log.New("release", "", l.LstdFlags, false)
	if err != nil {
		return
	}
	defer logger.Close()

	logger.Debug("will not print")
	logger.Release("My name is %v", name)

	log.Export(logger)

	log.Debug("will not print")
	log.Release("My name is %v", name)
}
