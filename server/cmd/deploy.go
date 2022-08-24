package cmd

import (
	"context"
	"fmt"
	"time"

	oapiclient "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	models "github.com/netlify/open-api/v2/go/models"
	netlify "github.com/netlify/open-api/v2/go/porcelain"
	ooapicontext "github.com/netlify/open-api/v2/go/porcelain/context"
	"github.com/robrotheram/gogallery/config"
	progressbar "github.com/schollz/progressbar/v3"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(deployCMD)
}

var deployCMD = &cobra.Command{
	Use:   "deploy",
	Short: "deploy static site",
	Long:  "deploy static site",
	Run: func(cmd *cobra.Command, args []string) {
		config := config.LoadConfig()
		config.Validate()
		fmt.Println("Deploying Site")
		deploySite(*config)
	},
}

func deploySite(c config.Configuration) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)
	client := netlify.NewRetryableHTTPClient(strfmt.NewFormats(), 10)
	ctx := ooapicontext.WithLogger(context.Background(), logger.WithFields(logrus.Fields{
		"source": "netlify",
	}))
	authCtx := ooapicontext.WithAuthInfo(ctx, oapiclient.BearerToken(c.Deploy.AuthToken))
	monitor := NewDeployMonitor()
	resp, err := client.DeploySite(authCtx, netlify.DeployOptions{
		SiteID:            c.Deploy.SiteId,
		Dir:               c.Gallery.Destpath,
		IsDraft:           c.Deploy.Draft,
		LargeMediaEnabled: true,
		Observer:          monitor,
		UploadTimeout:     20 * time.Minute,
		Title:             "gogallery deployment",
	})

	if err != nil {
		logger.Fatalf("failed to deploy site: %s", err)
	}
	monitor.Finish()

	// Print the site URL
	if resp.DeploySslURL != "" {
		fmt.Println("Site avalible: " + resp.DeploySslURL)
	} else if resp.DeployURL != "" {
		fmt.Println("Site avalible: " + resp.DeployURL)
	}
}

type DeployMonitor struct {
	bar *progressbar.ProgressBar
}

func NewDeployMonitor() *DeployMonitor {
	return &DeployMonitor{}
}

func (monitor *DeployMonitor) OnSetupWalk() error {
	return nil
}
func (monitor *DeployMonitor) OnSuccessfulStep(*netlify.FileBundle) error {
	return nil
}
func (monitor *DeployMonitor) OnSuccessfulWalk(files *models.DeployFiles) error {
	return nil
}
func (monitor *DeployMonitor) OnFailedWalk() {}
func (monitor *DeployMonitor) OnSetupDelta(bundle *models.DeployFiles) error {
	files, _ := bundle.Files.(map[string]string)
	size := len(files)
	monitor.bar = progressbar.Default(int64(size))
	return nil
}
func (monitor *DeployMonitor) OnSuccessfulDelta(bundle *models.DeployFiles, deploy *models.Deploy) error {
	return nil
}
func (monitor *DeployMonitor) OnFailedDelta(*models.DeployFiles) {}
func (monitor *DeployMonitor) OnSetupUpload(bundle *netlify.FileBundle) error {
	return nil
}
func (monitor *DeployMonitor) OnSuccessfulUpload(*netlify.FileBundle) error {
	monitor.bar.Add(1)
	return nil
}
func (monitor *DeployMonitor) OnFailedUpload(*netlify.FileBundle) {}
func (monitor *DeployMonitor) Finish()                            { monitor.bar.Finish() }
