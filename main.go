package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"

	"github.com/gkze/stars/starmanager"
	"github.com/urfave/cli"
)

func main() {
	sm, err := starmanager.New()
	if err != nil {
		log.Printf("Error creating StarManager! %v", err.Error())
	}

	cmdline := cli.NewApp()
	cmdline.Name = "stars"
	cmdline.Usage = "Command-line interface to YOUR GitHub stars"
	cmdline.Commands = []cli.Command{
		{
			Name:  "save",
			Usage: "Save all stars",
			Action: func(c *cli.Context) error {
				sm.SaveAllStars()

				return nil
			},
		},
		{
			Name:  "list-topics",
			Usage: "list all topics of starred projects",
			Action: func(c *cli.Context) error {
				sm.SaveIfEmpty()
				for _, pair := range sm.GetTopics() {
					fmt.Println(pair.Key, pair.Value)
				}

				return nil
			},
		},
		{
			Name:  "random",
			Usage: "Browse random stars",
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "count, c",
					Value: 6,
					Usage: "Number of random stars to browse",
				},
				cli.StringFlag{
					Name:  "language, l",
					Usage: "Limit to projects written only in this language",
				},
				cli.StringFlag{
					Name:  "topic, t",
					Usage: "Limit to projects with this topic",
				},
			},
			Action: func(c *cli.Context) error {
				sm.SaveIfEmpty()
				stars, err := sm.GetRandomProjects(c.Int("count"), c.String("language"), c.String("topic"))
				if err != nil {
					log.Printf(err.Error())
				}

				wg := sync.WaitGroup{}
				for _, proj := range stars {
					wg.Add(1)
					go func(p starmanager.Star) {
						defer wg.Done()
						cmd := exec.Command("/usr/bin/open", p.URL)
						err := cmd.Run()

						if err != nil {
							panic(err)
						}
					}(proj)
				}
				wg.Wait()

				return nil
			},
		},
		{
			Name:  "clear",
			Usage: "Clear local stars cache",
			Action: func(c *cli.Context) error {
				if err := sm.ClearCache(); err != nil {
					return err
				}

				return nil
			},
		},
		{
			Name:  "cleanup",
			Usage: "Clean up old stars",
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "months, m",
					Value: 2,
					Usage: "Number of months to delete projects older than",
				},
			},
			Action: func(c *cli.Context) error {
				sm.SaveIfEmpty()
				if err := sm.RemoveOlderThan(c.Int("months")); err != nil {
					return err
				}

				return nil
			},
		},
	}

	cmdline.Run(os.Args)
}
