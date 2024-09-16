package cmd

import (
	"encoding/json"
	"log"
	"os"

	"github.com/BESTSELLER/gcp-nuke/config"
	"github.com/BESTSELLER/gcp-nuke/gcp"
	"github.com/urfave/cli/v2"
)

// Command -
func Command() {
	app := &cli.App{
		Usage:     "The GCP project cleanup tool with added radiation",
		Version:   "v0.1.0",
		UsageText: "e.g. gcp-nuke --project test-nuke-262510 --dryrun",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "project, p",
				Usage:    "GCP project id to nuke (required)",
				Required: true,
			},
			&cli.BoolFlag{
				Name:  "dryrun, d",
				Usage: "Perform a dryrun instead",
			},
			&cli.IntFlag{
				Name:  "timeout, t",
				Value: 400,
				Usage: "Timeout for removal of a single resource in seconds",
			},
			&cli.IntFlag{
				Name:  "polltime, p",
				Value: 10,
				Usage: "Time for polling resource deletion status in seconds",
			},
			&cli.StringFlag{
				Name:    "exclusionsconfig, ec",
				Usage:   "Path to exclusions config file",
				EnvVars: []string{"EXCLUSIONS_CONFIG"},
				Aliases: []string{"ec"},
			},
			&cli.StringFlag{
				Name:    "gcpaccesstoken",
				Usage:   "GCP token for authentication",
				EnvVars: []string{"GCP_ACCESS_TOKEN"},
			},
		},
		Action: func(c *cli.Context) error {
			if c.String("gcpaccesstoken") == "" {
				return fmt.Errorf("GCP Access Token not provided")
			}
			token := config.ConvertStringToTokenSource(c.String("gcptoken"))
			// Behaviour to delete all resource in parallel in one project at a time - will be made into loop / concurrenct project nuke if required
			config := config.Config{
				Project:  c.String("project"),
				DryRun:   c.Bool("dryrun"),
				Timeout:  c.Int("timeout"),
				PollTime: c.Int("polltime"),
				Context:  gcp.Ctx,
				Zones:    gcp.GetZones(gcp.Ctx, c.String("project")),
				Regions:  gcp.GetRegions(gcp.Ctx, c.String("project")),
				GCPToken: token,
			}

			if c.String("exclusionsconfig") != "" {
				// Read exclusions config file and marshall into Config.Exclusions struct

				b, err := os.ReadFile(c.String("exclusionsconfig"))
				if err != nil {
					log.Printf("[Error] Exclusions config file not found at %v", c.String("exclusionsconfig"))
					return err
				}

				err = json.Unmarshal(b, &config.Exclusions)
				if err != nil {
					log.Printf("[Error] Exclusions config file could not be parsed")
				}

				log.Printf("Loaded exclusions config: %+v", config.Exclusions)
			}

			log.Printf("[Info] Timeout %v seconds. Polltime %v seconds. Dry run: %v", config.Timeout, config.PollTime, config.DryRun)
			gcp.RemoveProject(config)

			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
