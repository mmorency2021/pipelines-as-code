package pipelineascode

import (
	"regexp"
	"testing"

	apipac "github.com/openshift-pipelines/pipelines-as-code/pkg/apis/pipelinesascode/v1alpha1"
	"github.com/openshift-pipelines/pipelines-as-code/pkg/params"
	"github.com/openshift-pipelines/pipelines-as-code/pkg/params/clients"
	"github.com/openshift-pipelines/pipelines-as-code/pkg/params/info"
	kitesthelper "github.com/openshift-pipelines/pipelines-as-code/pkg/test/kubernetestint"
	"go.uber.org/zap"
	zapobserver "go.uber.org/zap/zaptest/observer"
	"gotest.tools/v3/assert"
	rtesting "knative.dev/pkg/reconciler/testing"
)

func TestSecretFromRepository(t *testing.T) {
	tests := []struct {
		name           string
		repo           *apipac.Repository
		providerconfig *info.ProviderConfig
		logmatch       []*regexp.Regexp
		expectedSecret string
		providerType   string
	}{
		{
			name: "config default",
			providerconfig: &info.ProviderConfig{
				APIURL: "https://apiurl.default",
			},
			expectedSecret: "configdefault",
			repo: &apipac.Repository{
				Spec: apipac.RepositorySpec{
					GitProvider: &apipac.GitProvider{
						Secret: &apipac.GitProviderSecret{
							Name: "repo-secret",
						},
					},
				},
			},
			providerType: "lalala",
			logmatch: []*regexp.Regexp{
				regexp.MustCompile("^Using git provider lalala: apiurl=https://apiurl.default user= token-secret=repo-secret in token-key=" + defaultGitProviderSecretKey),
			},
		},
		{
			name: "set api url",
			providerconfig: &info.ProviderConfig{
				APIURL: "https://donotwant",
			},
			repo: &apipac.Repository{
				Spec: apipac.RepositorySpec{
					GitProvider: &apipac.GitProvider{
						URL:    "https://dowant",
						Secret: &apipac.GitProviderSecret{},
					},
				},
			},
			expectedSecret: "setapiurl",
			logmatch: []*regexp.Regexp{
				regexp.MustCompile(".*apiurl=https://dowant.*"),
			},
		},
		{
			name:           "set user",
			providerconfig: &info.ProviderConfig{},
			repo: &apipac.Repository{
				Spec: apipac.RepositorySpec{
					GitProvider: &apipac.GitProvider{
						User:   "userfoo",
						Secret: &apipac.GitProviderSecret{},
					},
				},
			},
			expectedSecret: "set user",
			logmatch: []*regexp.Regexp{
				regexp.MustCompile(".*user=userfoo*"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, _ := rtesting.SetupFakeContext(t)
			observer, log := zapobserver.New(zap.InfoLevel)
			logger := zap.New(observer).Sugar()
			k8int := &kitesthelper.KinterfaceTest{
				GetSecretResult: tt.expectedSecret,
			}
			cs := &params.Run{
				Clients: clients.Clients{
					Log: logger,
				},
				Info: info.Info{
					Pac: &info.PacOpts{
						WebhookType: tt.providerType,
					},
				},
			}
			event := &info.Event{}

			err := secretFromRepository(ctx, cs, k8int, tt.providerconfig, event, tt.repo)
			assert.NilError(t, err)
			logs := log.TakeAll()
			assert.Equal(t, len(tt.logmatch), len(logs), "we didn't get the number of logging message: %+v", logs)
			for key, value := range logs {
				assert.Assert(t, tt.logmatch[key].MatchString(value.Message), "no match on logs %s => %s", tt.logmatch[key], value.Message)
			}
			assert.Assert(t, event.ProviderInfoFromRepo)
			assert.Equal(t, tt.expectedSecret, event.ProviderToken)
		})
	}
}
