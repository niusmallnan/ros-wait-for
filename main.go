package main

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/niusmallnan/ros-wait-for/check"
	"github.com/niusmallnan/ros-wait-for/types"
	"github.com/urfave/cli"
)

var VERSION = "v0.0.0-dev"

func main() {
	app := cli.NewApp()
	app.Name = "ros-wait-for"
	app.Version = VERSION
	app.Usage = "Waiting for something on RancherOS"
	app.Action = run
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:   "debug, d",
			EnvVar: "RANCHER_DEBUG",
		},
		cli.StringFlag{
			Name:   "containers",
			Usage:  "The name of containers",
			Value:  "",
			EnvVar: "RWF_CONTAINERS",
		},
		cli.StringFlag{
			Name:   "interfaces",
			Usage:  "The name of interfaces",
			Value:  "",
			EnvVar: "RWF_INTERFACES",
		},
		cli.DurationFlag{
			Name:   "timeout",
			Usage:  "Timeout duration for waiting",
			Value:  "10s",
			EnvVar: "RWF_TIMEOUT_DURATION",
		},
		cli.DurationFlag{
			Name:   "interval",
			Usage:  "Interval duration for checking",
			Value:  "1s",
			EnvVar: "RWF_INTERVAL_DURATION",
		},
	}
	app.Run(os.Args)
}

func run(c *cli.Context) error {
	if c.Bool("debug") {
		log.SetLevelString("debug")
	}

	checker, err := check.NewChecker(c.Duration("timeout"),
		c.Duration("interval"),
		c.String("containers"),
		c.String("interfaces"))
	if err != nil {
		logrus.Error(err)
		return err
	}
	exit := make(chan types.Exit)

	go func(exit chan<- types.Exit) {
		exit <- checker.Check()
	}(exit)

	go func(exit chan<- types.Exit) {
		exit <- checker.ThrowTimeout()
	}(exit)

	e := <-exit
	if e.Success {
		logrus.Info("Checking passed")
		return nil
	}
	logrus.Errorf("Exit with error: %v", e.Err)
	return e.Err
}
