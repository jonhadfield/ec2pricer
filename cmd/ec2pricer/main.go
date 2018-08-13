package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"github.com/jonhadfield/ec2pricer"
	"github.com/urfave/cli"
)

// overwritten at build time
var version, versionOutput, tag, sha, buildDate string

var (
	validOutputTypes = []string{"table", "yaml", "json"}
	locationsRegions = make(map[string]string)
	validRegions     = []string{
		"us-east-2", "us-east-1", "us-west-1", "us-west-2",
		"ap-south-1", "ap-northeast-2", "ap-northeast-3", "ap-southeast-1",
		"ap-southeast-2", "ap-northeast-1", "ca-central-1", "cn-north-1",
		"cn-northwest-1", "eu-central-1", "eu-west-1", "eu-west-2",
		"eu-west-3", "sa-east-1",
	}
	validLocations = []string{
		"US East (Ohio)", "US East (N. Virginia)", "US West (N. California)", "US West (Oregon)",
		"Asia Pacific (Mumbai)", "Asia Pacific (Seoul)", "Asia Pacific (Osaka-Local)", "Asia Pacific (Singapore)",
		"Asia Pacific (Sydney)", "Asia Pacific (Tokyo)", "Canada (Central)", "China (Beijing)",
		"China (Ningxia)", "EU (Frankfurt)", "EU (Ireland)", "EU (London)",
		"EU (Paris)", "South America (São Paulo)",
	}
)

func zipRegionsLocations() {
	if len(validLocations) != len(validRegions) {
		panic("must be equal number of regions and locations")
	}
	for x := 0; x < len(validLocations); x++ {
		locationsRegions[validLocations[x]] = validRegions[x]
	}
}

func main() {
	if tag != "" && buildDate != "" {
		versionOutput = fmt.Sprintf("[%s-%s] %s UTC", tag, sha, buildDate)
	} else {
		versionOutput = version
	}

	// read defaults from .ec2pricer file in current dir or home dir

	zipRegionsLocations()

	app := cli.NewApp()
	app.EnableBashCompletion = true

	app.Name = "ec2pricer"
	app.Version = versionOutput
	app.Compiled = time.Now()
	app.Authors = []cli.Author{
		{
			Name:  "Jon Hadfield",
			Email: "jon@lessknown.co.uk",
		},
	}
	app.HelpName = "-"
	app.Usage = "EC2 Pricer"
	app.Description = ""

	app.Commands = []cli.Command{
		{
			Name:  "instance",
			Usage: "get pricing for instances",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "type",
					Usage: "instance type (required)",
				},
				cli.StringFlag{
					Name:  "location",
					Usage: "instance location (required)",
				},
				cli.StringFlag{
					Name:  "os",
					Usage: "operating system",
				},
				cli.StringFlag{
					Name:  "tenancy",
					Usage: "dedicated or shared",
				},
				cli.StringFlag{
					Name:  "sw",
					Usage: "pre installed software",
				},
			},

			Action: func(c *cli.Context) error {
				instanceType := c.String("type")
				location := c.String("location")
				var validatedLocation string
				if instanceType == "" || location == "" {
					return cli.ShowCommandHelp(c, "instance")
				}

				// process location
				if ec2pricer.StringInSlice(location, validRegions, true) {
					validatedLocation = ec2pricer.GetKeyByVal(locationsRegions, location, true)
				}

				if ec2pricer.StringInSlice(location, validLocations, true) {
					validatedLocation = ec2pricer.GetMatchingKey(locationsRegions, location, true)
				}
				if validatedLocation == "" {
					log.Fatalf("location: \"%s\" does not exist", location)
				}

				appConfig := ec2pricer.InstanceAppConfig{
					InstanceType:    c.String("type"),
					Location:        validatedLocation,
					PreInstalledSw:  c.String("sw"),
					Tenancy:         c.String("tenancy"),
					OperatingSystem: c.String("os"),
				}
				ec2pricer.GetInstancePricing(&appConfig)
				return nil
			},
		},
	}

	app.Flags = []cli.Flag{
		cli.BoolFlag{Name: "debug"},
		cli.StringFlag{Name: "output, o"},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	if err := app.Run(os.Args); err != nil {
		os.Exit(1)
	}

}
