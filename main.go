package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
	"text/tabwriter"

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
	cmdline.Version = "0.4.3"
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
			Name:  "show",
			Usage: "Show popular stars given filters",
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "count, c",
					Value: 6,
					Usage: "Number of random stars to show",
				},
				cli.StringFlag{
					Name:  "language, l",
					Usage: "Limit to projects written only in this language",
				},
				cli.StringFlag{
					Name:  "topic, t",
					Usage: "Limit to projects with this topic",
				},
				cli.BoolFlag{
					Name:  "random, r",
					Usage: "Get random stars",
				},
				cli.BoolFlag{
					Name:  "browse, b",
					Usage: "Open stars in browser instead of writing them to stdout",
				},
			},
			Action: func(c *cli.Context) error {
				sm.SaveIfEmpty()

				stars, err := sm.GetProjects(
					c.Int("count"),
					c.String("language"),
					c.String("topic"),
					c.Bool("random"),
				)
				if err != nil {
					log.Printf(err.Error())
					return err
				}

				wg := sync.WaitGroup{}
				w := tabwriter.NewWriter(os.Stdout, 0, 8, 0, '\t', 0)

				for i := 0; i < len(stars); i++ {
					proj := stars[i]

					if c.Bool("browse") == true {
						wg.Add(1)

						go func(p starmanager.Star) {
							defer wg.Done()
							cmd := exec.Command("/usr/bin/open", p.URL)
							err := cmd.Run()

							if err != nil {
								panic(err)
							}
						}(proj)
					} else {
						if i == 0 {
							if c.String("language") == "" {
								fmt.Fprintf(w, "PUSHED\tSTARS\tLANGUAGE\tURL\tDESCRIPTION\n")
							} else {
								fmt.Fprintf(w, "PUSHED\tSTARS\tURL\tDESCRIPTION\n")
							}
						}

						if c.String("language") == "" {
							fmt.Fprintf(
								w,
								"%s\t%d\t%s\t%s\t%s\n",
								proj.PushedAt,
								proj.Stargazers,
								proj.Language,
								proj.URL,
								proj.Description,
							)
						} else {
							fmt.Fprintf(
								w,
								"%s\t%d\t%s\t%s\n",
								proj.PushedAt,
								proj.Stargazers,
								proj.URL,
								proj.Description,
							)
						}
					}
				}

				w.Flush()
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
				cli.BoolFlag{
					Name:  "include-archived, a",
					Usage: "Include archived stars",
				},
			},
			Action: func(c *cli.Context) error {
				sm.SaveIfEmpty()
				if err := sm.Cleanup(c.Int("months"), c.Bool("include-archived")); err != nil {
					return err
				}

				return nil
			},
		},
	}

	cmdline.Run(os.Args)
}
