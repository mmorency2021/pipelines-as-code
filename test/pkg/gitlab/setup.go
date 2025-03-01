package gitlab

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/openshift-pipelines/pipelines-as-code/pkg/params"
	"github.com/openshift-pipelines/pipelines-as-code/pkg/params/info"
	"github.com/openshift-pipelines/pipelines-as-code/pkg/provider/gitlab"
	"github.com/openshift-pipelines/pipelines-as-code/test/pkg/options"
	"github.com/openshift-pipelines/pipelines-as-code/test/pkg/repository"
	gitlab2 "github.com/xanzy/go-gitlab"
	"gotest.tools/v3/assert"
)

func Setup(ctx context.Context) (*params.Run, options.E2E, gitlab.Provider, error) {
	gitlabURL := os.Getenv("TEST_GITLAB_API_URL")
	gitlabToken := os.Getenv("TEST_GITLAB_TOKEN")
	sgitlabPid := os.Getenv("TEST_GITLAB_PROJECT_ID")
	gitlabPid, err := strconv.Atoi(sgitlabPid)
	if err != nil {
		return nil, options.E2E{}, gitlab.Provider{}, err
	}

	for _, value := range []string{
		"API_URL", "TOKEN", "PROJECT_ID",
	} {
		if env := os.Getenv("TEST_GITLAB_" + value); env == "" {
			return nil, options.E2E{}, gitlab.Provider{}, fmt.Errorf("\"TEST_%s\" env variable is required, cannot continue", value)
		}
	}
	run := &params.Run{}
	if err := run.Clients.NewClients(ctx, &run.Info); err != nil {
		return nil, options.E2E{}, gitlab.Provider{}, err
	}
	e2eoptions := options.E2E{
		ProjectID: gitlabPid,
	}
	glprovider := gitlab.Provider{}
	err = glprovider.SetClient(ctx,
		&info.Event{
			ProviderToken: gitlabToken,
			ProviderURL:   gitlabURL,
		},
	)
	if err != nil {
		return nil, options.E2E{}, gitlab.Provider{}, err
	}
	return run, e2eoptions, glprovider, nil
}

func TearDown(ctx context.Context, t *testing.T, runcnx *params.Run, glprovider gitlab.Provider, mrNumber int, ref string, targetNS string, projectid int) {
	runcnx.Clients.Log.Infof("Closing PR %d", mrNumber)
	if mrNumber != -1 {
		_, _, err := glprovider.Client.MergeRequests.UpdateMergeRequest(projectid, mrNumber,
			&gitlab2.UpdateMergeRequestOptions{StateEvent: gitlab2.String("close")})
		if err != nil {
			t.Fatal(err)
		}
	}
	repository.NSTearDown(ctx, t, runcnx, targetNS)
	runcnx.Clients.Log.Infof("Deleting Ref %s", ref)
	_, err := glprovider.Client.Branches.DeleteBranch(projectid, ref)
	assert.NilError(t, err)
}
