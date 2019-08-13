package leaf

import (
	"os"
	"os/signal"

	"github.com/adamluo159/leaf/cluster"
	"github.com/adamluo159/leaf/conf"
	"github.com/adamluo159/leaf/console"
	"github.com/adamluo159/leaf/jsonexcel"
	"github.com/adamluo159/leaf/log"
	"github.com/adamluo159/leaf/module"
)

func Run(mods ...module.Module) {
	// logger
	if conf.LogLevel != "" {
		logger, err := log.New(conf.LogLevel, conf.LogPath, conf.LogFlag, true)
		if err != nil {
			panic(err)
		}
		log.Export(logger)

		defer logger.Close()
	}

	jsonexcel.Init()

	log.Release("Leaf %v starting up", version)

	// module
	for i := 0; i < len(mods); i++ {
		module.Register(mods[i])
	}
	module.Init()

	// cluster
	cluster.Init()

	// console
	console.Init()

	// close
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	sig := <-c
	log.Release("Leaf closing down (signal: %v)", sig)
	console.Destroy()
	cluster.Destroy()
	module.Destroy()
}
