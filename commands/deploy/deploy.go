package deploy

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"path/filepath"
	"strings"

	"github.com/netlify/netlifyctl/commands/middleware"
	"github.com/netlify/netlifyctl/configuration"
	"github.com/netlify/netlifyctl/context"
	"github.com/netlify/netlifyctl/operations"
	netlify "github.com/netlify/open-api/go/porcelain"
	"github.com/spf13/cobra"
)

type deployCmd struct {
	base      string
	title     string
	draft     bool
	functions string
}

func Setup() (*cobra.Command, middleware.CommandFunc) {
	cmd := &deployCmd{}
	ccmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy your site",
		Long:  "Deploy your site",
	}
	ccmd.Flags().StringVarP(&cmd.base, "base-directory", "b", "", "directory to publish")
	ccmd.Flags().StringVarP(&cmd.title, "message", "m", "", "message for the deploy title")
	ccmd.Flags().BoolVarP(&cmd.draft, "draft", "d", false, "draft deploy, not published in production")
	ccmd.Flags().StringVarP(&cmd.functions, "functions", "f", "", "function directory to deploy")

	return ccmd, cmd.deploySite
}

func (dc *deployCmd) deploySite(ctx context.Context, cmd *cobra.Command, args []string) error {
	conf, err := middleware.ChooseSiteConf(ctx, cmd)
	if err != nil {
		return err
	}

	fmt.Println("=> Domain ready, deploying assets")

	client := context.GetClient(ctx)
	options := netlify.DeployOptions{
		SiteID: conf.Settings.ID,
		Dir:    baseDeploy(cmd, conf),
	}

	draft, err := cmd.Flags().GetBool("draft")
	if err != nil {
		logrus.WithError(err).Fatal("Failed to get string flag: 'draft'")
	}
	options.IsDraft = draft

	fs, err := cmd.Flags().GetString("functions")
	if err != nil {
		logrus.WithError(err).Fatal("Failed to get string flag: 'functions'")
	}
	options.FunctionsDir = fs

	logrus.WithFields(logrus.Fields{
		"site":  options.SiteID,
		"path":  options.Dir,
		"draft": options.IsDraft}).Debug("deploying site")

	d, err := client.DeploySite(ctx, options)
	if err != nil {
		return err
	}

	if len(d.Required) > 0 {
		ready, err := client.WaitUntilDeployReady(ctx, d)
		if err != nil {
			return err
		}
		d = ready
	}
	fmt.Printf("=> Done, your website is live in %s\n", d.URL)

	return nil
}

func baseDeploy(cmd *cobra.Command, conf *configuration.Configuration) string {
	bd, err := cmd.Flags().GetString("base-directory")
	if err != nil {
		logrus.WithError(err).Fatal("Failed to get string flag: 'base-directory'")
	}

	if bd != "" {
		return bd
	}
	s := conf.Settings
	var path = s.Path
	if path == "" {
		path, err = operations.AskForInput("What path would you like deployed?", ".")
		if err != nil {
			logrus.WithError(err).Fatal("Failed to get deploy path")
		}

		s.Path = path
		logrus.Debugf("Got new path from the user %s", s.Path)
	}

	if !strings.HasPrefix(s.Path, "/") {
		path = filepath.Join(conf.Root(), s.Path)
		logrus.Debugf("Relative path detected, going to deploy: '%s'", path)
	}

	return path
}
