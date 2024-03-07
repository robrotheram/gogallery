package deploy

import (
	"context"
	"fmt"
	"time"

	oapiclient "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	models "github.com/netlify/open-api/v2/go/models"
	porcelain "github.com/netlify/open-api/v2/go/porcelain"
	ooapicontext "github.com/netlify/open-api/v2/go/porcelain/context"
	"github.com/robrotheram/gogallery/backend/config"
	"github.com/robrotheram/gogallery/backend/monitor"
	"github.com/sirupsen/logrus"
)

func DeploySite(c config.Configuration, stats *monitor.ProgressStats) error {
	if len(c.Deploy.SiteId) == 0 || len(c.Deploy.AuthToken) == 0 {
		return fmt.Errorf("no deployment config found")
	}
	stats.Start()
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)
	client := porcelain.NewRetryableHTTPClient(strfmt.NewFormats(), 10)
	ctx := ooapicontext.WithLogger(context.Background(), logger.WithFields(logrus.Fields{
		"source": "netlify",
	}))
	authCtx := ooapicontext.WithAuthInfo(ctx, oapiclient.BearerToken(c.Deploy.AuthToken))

	obs := NewDeployObserver(stats)

	resp, err := client.DeploySite(authCtx, porcelain.DeployOptions{
		SiteID:            c.Deploy.SiteId,
		Observer:          obs,
		Dir:               c.Gallery.Destpath,
		IsDraft:           c.Deploy.Draft,
		LargeMediaEnabled: true,
		UploadTimeout:     20 * time.Minute,
		Title:             "gogallery deployment",
	})

	if err != nil {
		return fmt.Errorf("failed to deploy site: %s", err)
	}
	// Print the site URL
	if resp.DeploySslURL != "" {
		fmt.Println("Site avalible: " + resp.DeploySslURL)
	} else if resp.DeployURL != "" {
		fmt.Println("Site avalible: " + resp.DeployURL)
	}
	stats.End()
	return nil
}

type DeployObserver struct {
	stats *monitor.ProgressStats
}

func NewDeployObserver(stats *monitor.ProgressStats) *DeployObserver {
	return &DeployObserver{
		stats: stats,
	}
}
func (o *DeployObserver) OnSetupWalk() error {
	o.stats.Update()
	return nil
}

func (o *DeployObserver) OnSuccessfulStep(*porcelain.FileBundle) error {
	o.stats.Update()
	return nil
}

func (o *DeployObserver) OnSuccessfulWalk(df *models.DeployFiles) error {
	o.stats.Update()
	return nil
}

func (o *DeployObserver) OnFailedWalk() {
	o.stats.Update()
}

func (o *DeployObserver) OnSetupDelta(*models.DeployFiles) error {
	o.stats.Update()
	return nil
}

func (o *DeployObserver) OnSuccessfulDelta(df *models.DeployFiles, d *models.Deploy) error {
	o.stats.Update()
	return nil
}

func (o *DeployObserver) OnFailedDelta(*models.DeployFiles) {
	o.stats.Update()
}

func (o *DeployObserver) OnSetupUpload(f *porcelain.FileBundle) error {
	return nil
}

func (o *DeployObserver) OnSuccessfulUpload(f *porcelain.FileBundle) error {
	o.stats.Update()
	return nil
}

func (o *DeployObserver) OnFailedUpload(*porcelain.FileBundle) {
	o.stats.Update()
}

func (o *DeployObserver) Finish() {
	o.stats.Update()
}
