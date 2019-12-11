package main

import (
	"fmt"
	"log"
	"os"
	"sync"
	"text/tabwriter"

	"github.com/gkze/stars/starmanager"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

// Version is version information dynamically injected at build time
var Version string

func main() {
	sm, err := starmanager.New()
	if err != nil {
		log.Printf("Error creating StarManager! %v", err.Error())
	}

	starsCmd := &cobra.Command{
		Use:   "stars",
		Short: "Stars is a command-line GitHub Stars manager",
		Long: `A CLI written in Golang to facilitate efficient management of a user's
GitHub starred projects / repositories, a.k.a. "Stars"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Help()
			return nil
		},
	}

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show version of stars",
		Long:  "Displays the version of the currently running stars CLI binary",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("stars version %s\n", Version)
			return nil
		},
	}

	saveAllStarsCmd := &cobra.Command{
		Use:   "save",
		Short: "Save all stars",
		Long:  "Fetches all of the current user's starred projects to the local filesystem",
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := sm.SaveAllStars(); err != nil {
				return err
			}

			return nil
		},
	}

	topicsCmd := &cobra.Command{
		Use:   "topics",
		Short: "List all topics of all stars",
		Long:  "Displays a list of topics, sorted by occurrece count, for all of a user's starred projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := sm.SaveIfEmpty(); err != nil {
				return err
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 8, 0, '\t', 0)

			for i, pair := range sm.GetTopics() {
				if i == 0 {
					fmt.Fprintf(w, "TOPIC\tOCCURRENCES\n")
				}

				fmt.Fprintf(w, "%s\t%d\n", pair.Key, pair.Value)
			}

			return w.Flush()
		},
	}

	var (
		count    int
		language string
		topic    string
		random   bool
		browse   bool
	)

	showStarsCmd := &cobra.Command{
		Use:   "show",
		Short: "Show stars",
		Long:  "Displays a tabulated list of stars given query parameters",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := sm.SaveIfEmpty(); err != nil {
				return err
			}

			stars, err := sm.GetProjects(count, language, topic, random)
			if err != nil {
				log.Printf(err.Error())
				return err
			}

			wg := sync.WaitGroup{}
			w := tabwriter.NewWriter(os.Stdout, 0, 8, 0, '\t', 0)

			for i := 0; i < len(stars); i++ {
				proj := stars[i]

				if browse {
					wg.Add(1)

					go func(p starmanager.Star) {
						defer wg.Done()
						if err := browser.OpenURL(p.URL); err != nil {
							panic(err)
						}
					}(proj)
				} else {
					if i == 0 {
						if language == "" {
							fmt.Fprintf(w, "PUSHED\tSTARS\tLANGUAGE\tURL\tDESCRIPTION\n")
						} else {
							fmt.Fprintf(w, "PUSHED\tSTARS\tURL\tDESCRIPTION\n")
						}
					}

					if language == "" {
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

			if err := w.Flush(); err != nil {
				return err
			}

			wg.Wait()

			return nil
		},
	}

	showStarsCmd.PersistentFlags().IntVarP(&count, "count", "c", 6, "Number of stars to show")
	showStarsCmd.PersistentFlags().StringVarP(&language, "language", "l", "", "Limit to projects written only in this language")
	showStarsCmd.PersistentFlags().StringVarP(&topic, "topic", "t", "", "Limit to projects with this topic")
	showStarsCmd.PersistentFlags().BoolVarP(&random, "random", "r", false, "Randomize results")
	showStarsCmd.PersistentFlags().BoolVarP(&browse, "browse", "b", false, "Open stars in browser instead of writing them to stdout")

	clearCmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear local stars cache",
		Long:  "Wipe the file on the local filesystem containing the fetched results of all stars",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := sm.ClearCache(); err != nil {
				return err
			}

			return nil
		},
	}

	var (
		months          int
		includeArchived bool
	)

	cleanupCmd := &cobra.Command{
		Use:   "cleanup",
		Short: "Clean up old stars",
		Long:  "Un-stars projects older than n months, optionally also unstarring archived projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := sm.SaveIfEmpty(); err != nil {
				return err
			}

			if err := sm.Cleanup(months, includeArchived); err != nil {
				return err
			}

			return nil
		},
	}

	cleanupCmd.PersistentFlags().IntVarP(&months, "months", "m", 2, "Number of months to delete projects older than")
	cleanupCmd.PersistentFlags().BoolVarP(&includeArchived, "include-archived", "a", false, "Include archived stars")

	completionCmd := &cobra.Command{
		Use:   "completion",
		Short: "Generate completion",
		Long:  "Outputs an autocompletion script to be sourced by a target shell",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(`Outputs autocompletion scripts for the CLI. Please refer
to your shell's documentation on how to configure autocompletion.
			`)
			cmd.Help()

			return nil
		},
	}

	bashCompletionCmd := &cobra.Command{
		Use:   "bash",
		Short: "Generate bash completion",
		Long:  "Outputs Bash autocompletion script",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.GenBashCompletion(os.Stdout)
		},
	}

	zshCompletionCmd := &cobra.Command{
		Use:   "zsh",
		Short: "Generate Zsh completion",
		Long:  "Outputs Zsh autocompletion script",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.GenZshCompletion(os.Stdout)
		},
	}

	completionCmd.AddCommand(bashCompletionCmd, zshCompletionCmd)

	starsCmd.AddCommand(
		versionCmd,
		saveAllStarsCmd,
		topicsCmd,
		showStarsCmd,
		clearCmd,
		cleanupCmd,
		completionCmd,
	)

	if err := starsCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
