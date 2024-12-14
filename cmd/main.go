package main

import (
	"context"
	"os"

	"github.com/arhyth/mitch"
	"github.com/arhyth/mitch/internal"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:           "mitch",
		Usage:          "A simple migration tool for Clickhouse",
		DefaultCommand: "run",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "env",
				Aliases: []string{"e"},
				Usage:   "Path to the .env file",
			},
			&cli.StringFlag{
				Name:  "rollback",
				Usage: "Path to the SQL rollback file (optional, triggers rollback mode)",
			},
		},
		Commands: []*cli.Command{
			{
				Name: "testhelper",
				Args: true,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "db-name",
						Aliases:  []string{"db"},
						Usage:    "Name of temporary DB to be created for testing",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "tempdir",
						Aliases:  []string{"td"},
						Usage:    "Temporary directory for binary and env file",
						Required: true,
					},
				},
				Subcommands: []*cli.Command{
					{
						Name:  "create",
						Usage: "Create temporary DB for testing",
						Action: func(c *cli.Context) error {
							dbname := c.String("db-name")
							help, err := NewDBHelper()
							if err != nil {
								return err
							}

							if err = help.CreateDatabase(dbname); err != nil {
								return err
							}

							tempdir := c.String("tempdir")
							createEnvFile(tempdir, dbname)

							return nil
						},
					},
					{
						Name:  "drop",
						Usage: "Drop temporary DB",
						Action: func(c *cli.Context) error {
							dbname := c.String("db-name")
							help, err := NewDBHelper()
							if err != nil {
								return err
							}

							if err = help.DropDatabase(dbname); err != nil {
								return err
							}

							return nil
						},
					},
				},
			},
			{
				Name: "run",
				Args: true,
				Action: func(c *cli.Context) error {
					configPath := c.String("env")
					log.Info().
						Str("path", configPath).
						Msg("loading env vars from file...")
					if err := godotenv.Load(configPath); err != nil {
						return err
					}

					rollbackFile := c.String("rollback")
					dbURL := c.Args().First()
					migrationDir := os.Getenv(mitch.EnvMigrationDir)
					if migrationDir == "" {
						log.Warn().Msgf(
							"Migration directory env `%s` not set, defaulting to `%s`",
							mitch.EnvMigrationDir,
							mitch.DefaultMigrationDir,
						)
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
					runner := internal.NewRunner(dirFs, conn)

					// Rollback mode
					if rollbackFile != "" {
						log.Debug().
							Str("file", rollbackFile).
							Msg("Running in rollback mode...")
						if err = runner.Rollback(context.Background(), rollbackFile); err != nil {
							return err
						}
						return nil
					}

					// Forward mode
					log.Debug().Msg("Running in forward mode...")
					if err = runner.Migrate(context.Background()); err != nil {
						log.Error().Err(err).Msg("runner.Migrate failed")
						return err
					}

					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("failed to run command")
	}
}
