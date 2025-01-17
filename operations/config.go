package operations

import (
	"context"

	"github.com/evergreen-ci/cedar"
	"github.com/evergreen-ci/cedar/model"
	"github.com/evergreen-ci/utility"
	"github.com/mongodb/grip"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func dumpCedarConfig() cli.Command {
	return cli.Command{
		Name:  "dump",
		Usage: "write current Cedar application configuration to a file",
		Flags: dbFlags(
			cli.StringFlag{
				Name:  "file",
				Usage: "specify path to a Cedar application config file",
			}),
		Action: func(c *cli.Context) error {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			fileName := c.String("file")
			mongodbURI := c.String(dbURIFlag)
			dbName := c.String(dbNameFlag)
			dbCredFile := c.String(dbCredsFileFlag)

			sc := newServiceConf(2, true, mongodbURI, "", dbName, dbCredFile)
			sc.interactive = true

			if err := sc.setup(ctx); err != nil {
				return errors.WithStack(err)
			}

			env := cedar.GetEnvironment()
			conf := &model.CedarConfig{}
			conf.Setup(env)

			if err := conf.Find(); err != nil {
				return errors.WithStack(err)
			}

			return errors.WithStack(utility.WriteJSONFile(fileName, conf))
		},
	}
}

func loadCedarConfig() cli.Command {
	return cli.Command{
		Name:  "load",
		Usage: "loads Cedar application configuration from a file",
		Flags: dbFlags(
			cli.StringFlag{
				Name:  "file",
				Usage: "specify path to a Cedar application config file",
			}),
		Action: func(c *cli.Context) error {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			fileName := c.String("file")
			mongodbURI := c.String(dbURIFlag)
			dbName := c.String(dbNameFlag)
			dbCredFile := c.String(dbCredsFileFlag)

			conf, err := model.LoadCedarConfig(fileName)
			if err != nil {
				return errors.WithStack(err)
			}

			sc := newServiceConf(2, true, mongodbURI, "", dbName, dbCredFile)
			sc.interactive = true
			if err = sc.setup(ctx); err != nil {
				return errors.WithStack(err)
			}
			env := cedar.GetEnvironment()

			conf.Setup(env)
			if err = conf.Save(); err != nil {
				return errors.WithStack(err)
			}

			grip.Infoln("successfully application configuration to DB at:", mongodbURI)
			return nil
		},
	}
}
