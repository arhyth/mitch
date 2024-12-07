package main

import (
	"io/fs"
	"os"

	"github.com/arhyth/mitch"
	"github.com/arhyth/mitch/internal"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "mitch",
		Usage: "A simple migration tool for Clickhouse",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "config",
				Aliases:  []string{"c"},
				Usage:    "Path to the configuration file",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "rollback",
				Usage: "Path to the SQL rollback file (optional, triggers rollback mode)",
			},
		},
		Before: func(c *cli.Context) error {
			configPath := c.String("config")
			log.Info().
				Str("path", configPath).
				Msg("loading env vars from config file...")
			if err := godotenv.Load(configPath); err != nil {
				return err
			}
			return nil
		},
		Action: func(c *cli.Context) error {
			rollbackFile := c.String("rollback")
			dbURL := c.Args().First()
			migrationDir := c.String(mitch.EnvMigrationDir)
			if migrationDir != "" {
				log.Warn().Msgf(
					"Migration directory env `%s` not set, defaulting to `%s`",
					mitch.EnvMigrationDir,
					mitch.DefaultMigrationDir,
				)
			} else {
				migrationDir = mitch.DefaultMigrationDir
			}

			if dbURL == "" {
				log.Warn().Msg("DB URL argument not set, will try env vars...")
			}

			conn, err := mitch.Connect(dbURL)
			if err != nil {
				return err
			}
			dirFs := os.DirFS(migrationDir)
			if sf, ok := dirFs.(fs.StatFS); ok {
				if _, err = sf.Stat("."); err != nil {
					return err
				}
			}

			runner := internal.Runner{
				Dir: dirFs,
				DB:  conn,
			}

			// Rollback mode
			if rollbackFile != "" {
				log.Info().
					Str("file", rollbackFile).
					Msg("Running in rollback mode...")
				if err = runner.RollbackTo(rollbackFile); err != nil {
					return err
				}
				return nil
			}

			// Forward mode
			log.Info().Msg("Running in forward mode...")
			if err = runner.Migrate(); err != nil {
				return err
			}

			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("failed to run command")
	}
}
