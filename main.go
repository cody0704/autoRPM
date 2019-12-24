package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
	"github.com/urfave/cli"
)

// Version set at compile-time
var Version string

func main() {
	// Load env-file if it exists first
	if filename, found := os.LookupEnv("PLUGIN_ENV_FILE"); found {
		_ = godotenv.Load(filename)
	}

	app := cli.NewApp()
	app.Name = "Auto RPM"
	app.Usage = "Executing rpm packaging"
	app.Copyright = "Copyright (c) 2019 Cody Chen"
	app.Authors = []cli.Author{
		{
			Name:  "Cody Chen",
			Email: "cody@acom-networks.com",
		},
	}
	app.Action = run
	app.Version = Version
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "rpm.name",
			Usage:  "rpm filename",
			EnvVar: "RPM_NAME,INPUT_RPM_NAME",
		},
		cli.StringFlag{
			Name:   "rpm.path",
			Usage:  "package path",
			EnvVar: "RPM_PATH,INPUT_RPM_PATH",
		},
		cli.StringFlag{
			Name:   "rpm.packager",
			Usage:  "rpm packager",
			EnvVar: "RPM_PACKAGER,INPUT_PACKAGER",
		},
		cli.StringFlag{
			Name:   "rpm.packaging.path",
			Usage:  "rpm packaging path",
			EnvVar: "RPM_PACKAGING_PATH,INPUT_PACKAGING_PATH",
		},
		cli.StringFlag{
			Name:   "rpm.project.name",
			Usage:  "rpm project name",
			EnvVar: "RPM_PROJECT_NAME,INPUT_PROJECT_NAME",
		},
		cli.StringFlag{
			Name:   "rpm.url",
			Usage:  "rpm url",
			EnvVar: "RPM_URL,INPUT_RPM_URL",
		},
		cli.StringFlag{
			Name:   "rpm.vendor",
			Usage:  "rpm vendor",
			EnvVar: "RPM_VENDOR,INPUT_RPM_VENDOR",
		},
		cli.StringFlag{
			Name:   "rpm.apiname",
			Usage:  "rpm apiname",
			EnvVar: "RPM_APINAME,INPUT_RPM_APINAME",
		},
		cli.StringFlag{
			Name:   "rpm.version",
			Usage:  "rpm version",
			EnvVar: "RPM_VERSION,INPUT_RPM_VERSION",
		},
		cli.StringFlag{
			Name:   "rpm.requires",
			Usage:  "rpm requires",
			EnvVar: "RPM_REQUIRES,INPUT_RPM_REQUIRES",
		},
		cli.StringFlag{
			Name:   "git.enable",
			Usage:  "git token",
			EnvVar: "GIT_ENABLE,INPUT_GIT_ENABLE",
		},
		cli.StringFlag{
			Name:   "reponstiry",
			Usage:  "execute single commands for github action",
			EnvVar: "RPM_REPONSTIRY,INPUT_REPONSTIRY",
		},
		cli.StringFlag{
			Name:   "git.token",
			Usage:  "git token",
			EnvVar: "GIT_TOKEN,INPUT_GIT_TOKEN",
		},
	}

	// Override a template
	cli.AppHelpTemplate = `
    _         _        ____  ____  __  __
   / \  _   _| |_ ___ |  _ \|  _ \|  \/  |
  / _ \| | | | __/ _ \| |_) | |_) | |\/| |
 / ___ \ |_| | || (_) |  _ <|  __/| |  | |
/_/   \_\__,_|\__\___/|_| \_\_|   |_|  |_|
                                                    version: {{.Version}}
NAME:
   {{.Name}} - {{.Usage}}
AUTHOR:
   {{range .Authors}}{{ . }}{{end}}
   {{end}}{{if .Commands}}
COPYRIGHT:
   {{.Copyright}}
   {{end}}{{if .Version}}
VERSION:
   {{.Version}}
   {{end}}
REPOSITORY:
    Github: https://github.com/cody0704/autoRPM
`

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(c *cli.Context) error {
	reponstiry := c.String("reponstiry")

	plugin := Plugin{
		Config: Config{
			RPMNAME:       c.String("rpm.name"),
			PackagePath:   c.String("rpm.path"),
			Packager:      c.String("rpm.packager"),
			PackagingPath: c.String("rpm.packaging.path"),
			ProjectName:   c.String("rpm.project.name"),
			URL:           c.String("rpm.url"),
			Vendor:        c.String("rpm.vendor"),
			APINAME:       c.String("rpm.apiname"),
			Version:       c.String("rpm.version"),
			Requires:      c.String("rpm.requires"),
			GitEnable:     c.Bool("git.enable"),
			GitToken:      c.String("git.token"),
			Reponstiry:    reponstiry,
		},
		Writer: os.Stdout,
	}

	return plugin.Exec()
}
