package deployconfig

import (
	"fmt"
	"strings"
	"testing"
	"time"

	kubeapi "github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/labels"
	"github.com/openshift/origin/pkg/deploy/api"
	"github.com/openshift/origin/pkg/deploy/registry/test"
)

func TestListDeploymentConfigsError(t *testing.T) {
	mockRegistry := test.NewDeploymentConfigRegistry()
	mockRegistry.Err = fmt.Errorf("test error")

	storage := REST{
		registry: mockRegistry,
	}

	deploymentConfigs, err := storage.List(nil, nil, nil)
	if err != mockRegistry.Err {
		t.Errorf("Expected %#v, Got %#v", mockRegistry.Err, err)
	}

	if deploymentConfigs != nil {
		t.Errorf("Unexpected non-nil deploymentConfigs list: %#v", deploymentConfigs)
	}
}

func TestListDeploymentConfigsEmptyList(t *testing.T) {
	mockRegistry := test.NewDeploymentConfigRegistry()
	mockRegistry.DeploymentConfigs = &api.DeploymentConfigList{
		Items: []api.DeploymentConfig{},
	}

	storage := REST{
		registry: mockRegistry,
	}

	deploymentConfigs, err := storage.List(nil, labels.Everything(), labels.Everything())
	if err != nil {
		t.Errorf("Unexpected non-nil error: %#v", err)
	}

	if len(deploymentConfigs.(*api.DeploymentConfigList).Items) != 0 {
		t.Errorf("Unexpected non-zero deploymentConfigs list: %#v", deploymentConfigs)
	}
}

func TestListDeploymentConfigsPopulatedList(t *testing.T) {
	mockRegistry := test.NewDeploymentConfigRegistry()
	mockRegistry.DeploymentConfigs = &api.DeploymentConfigList{
		Items: []api.DeploymentConfig{
			{
				JSONBase: kubeapi.JSONBase{
					ID: "foo",
				},
			},
			{
				JSONBase: kubeapi.JSONBase{
					ID: "bar",
				},
			},
		},
	}

	storage := REST{
		registry: mockRegistry,
	}

	list, err := storage.List(nil, labels.Everything(), labels.Everything())
	if err != nil {
		t.Errorf("Unexpected non-nil error: %#v", err)
	}

	deploymentConfigs := list.(*api.DeploymentConfigList)

	if e, a := 2, len(deploymentConfigs.Items); e != a {
		t.Errorf("Expected %v, got %v", e, a)
	}
}

func TestCreateDeploymentConfigBadObject(t *testing.T) {
	storage := REST{}

	channel, err := storage.Create(nil, &api.DeploymentList{})
	if channel != nil {
		t.Errorf("Expected nil, got %v", channel)
	}
	if strings.Index(err.Error(), "not a deploymentConfig") == -1 {
		t.Errorf("Expected 'not a deploymentConfig' error, got '%v'", err.Error())
	}
}

func TestCreateRegistrySaveError(t *testing.T) {
	mockRegistry := test.NewDeploymentConfigRegistry()
	mockRegistry.Err = fmt.Errorf("test error")
	storage := REST{registry: mockRegistry}

	channel, err := storage.Create(nil, &api.DeploymentConfig{
		JSONBase: kubeapi.JSONBase{ID: "foo"},
	})
	if channel == nil {
		t.Errorf("Expected nil channel, got %v", channel)
	}
	if err != nil {
		t.Errorf("Unexpected non-nil error: %#v", err)
	}

	select {
	case result := <-channel:
		status, ok := result.(*kubeapi.Status)
		if !ok {
			t.Errorf("Expected status type, got: %#v", result)
		}
		if status.Status != kubeapi.StatusFailure || status.Message != "foo" {
			t.Errorf("Expected failure status, got %#V", status)
		}
	case <-time.After(50 * time.Millisecond):
		t.Errorf("Timed out waiting for result")
	default:
	}
}

func TestCreateDeploymentConfigOK(t *testing.T) {
	mockRegistry := test.NewDeploymentConfigRegistry()
	storage := REST{registry: mockRegistry}

	channel, err := storage.Create(nil, &api.DeploymentConfig{
		JSONBase: kubeapi.JSONBase{ID: "foo"},
	})
	if channel == nil {
		t.Errorf("Expected nil channel, got %v", channel)
	}
	if err != nil {
		t.Errorf("Unexpected non-nil error: %#v", err)
	}

	select {
	case result := <-channel:
		deploymentConfig, ok := result.(*api.DeploymentConfig)
		if !ok {
			t.Errorf("Expected deploymentConfig type, got: %#v", result)
		}
		if deploymentConfig.ID != "foo" {
			t.Errorf("Unexpected deploymentConfig: %#v", deploymentConfig)
		}
	case <-time.After(50 * time.Millisecond):
		t.Errorf("Timed out waiting for result")
	default:
	}
}

func TestGetDeploymentConfigError(t *testing.T) {
	mockRegistry := test.NewDeploymentConfigRegistry()
	mockRegistry.Err = fmt.Errorf("bad")
	storage := REST{registry: mockRegistry}

	deploymentConfig, err := storage.Get(nil, "foo")
	if deploymentConfig != nil {
		t.Errorf("Unexpected non-nil deploymentConfig: %#v", deploymentConfig)
	}
	if err != mockRegistry.Err {
		t.Errorf("Expected %#v, got %#v", mockRegistry.Err, err)
	}
}

func TestGetDeploymentConfigOK(t *testing.T) {
	mockRegistry := test.NewDeploymentConfigRegistry()
	mockRegistry.DeploymentConfig = &api.DeploymentConfig{
		JSONBase: kubeapi.JSONBase{ID: "foo"},
	}
	storage := REST{registry: mockRegistry}

	deploymentConfig, err := storage.Get(nil, "foo")
	if deploymentConfig == nil {
		t.Error("Unexpected nil deploymentConfig")
	}
	if err != nil {
		t.Errorf("Unexpected non-nil error", err)
	}
	if deploymentConfig.(*api.DeploymentConfig).ID != "foo" {
		t.Errorf("Unexpected deploymentConfig: %#v", deploymentConfig)
	}
}

func TestUpdateDeploymentConfigBadObject(t *testing.T) {
	storage := REST{}

	channel, err := storage.Update(nil, &api.DeploymentList{})
	if channel != nil {
		t.Errorf("Expected nil, got %v", channel)
	}
	if strings.Index(err.Error(), "not a deploymentConfig:") == -1 {
		t.Errorf("Expected 'not a deploymentConfig' error, got %v", err)
	}
}

func TestUpdateDeploymentConfigMissingID(t *testing.T) {
	storage := REST{}

	channel, err := storage.Update(nil, &api.DeploymentConfig{})
	if channel != nil {
		t.Errorf("Expected nil, got %v", channel)
	}
	if strings.Index(err.Error(), "id is unspecified:") == -1 {
		t.Errorf("Expected 'id is unspecified' error, got %v", err)
	}
}

func TestUpdateRegistryErrorSaving(t *testing.T) {
	mockRepositoryRegistry := test.NewDeploymentConfigRegistry()
	mockRepositoryRegistry.Err = fmt.Errorf("foo")
	storage := REST{registry: mockRepositoryRegistry}

	channel, err := storage.Update(nil, &api.DeploymentConfig{
		JSONBase: kubeapi.JSONBase{ID: "bar"},
	})
	if err != nil {
		t.Errorf("Unexpected non-nil error: %#v", err)
	}
	result := <-channel
	status, ok := result.(*kubeapi.Status)
	if !ok {
		t.Errorf("Expected status, got %#v", result)
	}
	if status.Status != kubeapi.StatusFailure || status.Message != "foo" {
		t.Errorf("Expected status=failure, message=foo, got %#v", status)
	}
}

func TestUpdateDeploymentConfigOK(t *testing.T) {
	mockRepositoryRegistry := test.NewDeploymentConfigRegistry()
	storage := REST{registry: mockRepositoryRegistry}

	channel, err := storage.Update(nil, &api.DeploymentConfig{
		JSONBase: kubeapi.JSONBase{ID: "bar"},
	})
	if err != nil {
		t.Errorf("Unexpected non-nil error: %#v", err)
	}
	result := <-channel
	repo, ok := result.(*api.DeploymentConfig)
	if !ok {
		t.Errorf("Expected DeploymentConfig, got %#v", result)
	}
	if repo.ID != "bar" {
		t.Errorf("Unexpected repo returned: %#v", repo)
	}
}

func TestDeleteDeploymentConfig(t *testing.T) {
	mockRegistry := test.NewDeploymentConfigRegistry()
	storage := REST{registry: mockRegistry}
	channel, err := storage.Delete(nil, "foo")
	if channel == nil {
		t.Error("Unexpected nil channel")
	}
	if err != nil {
		t.Errorf("Unexpected non-nil error: %#v", err)
	}

	select {
	case result := <-channel:
		status, ok := result.(*kubeapi.Status)
		if !ok {
			t.Errorf("Expected status type, got: %#v", result)
		}
		if status.Status != kubeapi.StatusSuccess {
			t.Errorf("Expected status=success, got: %#v", status)
		}
	case <-time.After(50 * time.Millisecond):
		t.Errorf("Timed out waiting for result")
	default:
	}
}
