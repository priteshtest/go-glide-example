package project

import (
	"fmt"

	kubeapi "github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/api/errors"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/apiserver"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/labels"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/runtime"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/util"

	"github.com/openshift/origin/pkg/project/api"
	"github.com/openshift/origin/pkg/project/api/validation"
)

// REST implements the RESTStorage interface in terms of an Registry.
type REST struct {
	registry Registry
}

// NewStorage returns a new REST.
func NewREST(registry Registry) apiserver.RESTStorage {
	return &REST{registry}
}

// New returns a new Project for use with Create and Update.
func (s *REST) New() runtime.Object {
	return &api.Project{}
}

// List retrieves a list of Projects that match selector.
func (s *REST) List(ctx kubeapi.Context, selector, fields labels.Selector) (runtime.Object, error) {
	projects, err := s.registry.ListProjects(ctx, selector)
	if err != nil {
		return nil, err
	}

	return projects, nil
}

// Get retrieves an Project by id.
func (s *REST) Get(ctx kubeapi.Context, id string) (runtime.Object, error) {
	project, err := s.registry.GetProject(ctx, id)
	if err != nil {
		return nil, err
	}
	return project, nil
}

// Create registers the given Project.
func (s *REST) Create(ctx kubeapi.Context, obj runtime.Object) (<-chan runtime.Object, error) {
	project, ok := obj.(*api.Project)
	if !ok {
		return nil, fmt.Errorf("not a project: %#v", obj)
	}

	// TODO decide if we should set namespace == name, think longer term we need some type of reservation here
	// but i want to be able to let existing kubernetes ns grow into a project as well
	if len(project.Namespace) == 0 {
		project.Namespace = project.ID
	}

	// TODO set an id if not provided?, set a Namespace attribute if not provided?
	project.CreationTimestamp = util.Now()

	if errs := validation.ValidateProject(project); len(errs) > 0 {
		return nil, errors.NewInvalid("project", project.ID, errs)
	}

	return apiserver.MakeAsync(func() (runtime.Object, error) {
		if err := s.registry.CreateProject(ctx, project); err != nil {
			return nil, err
		}
		return s.Get(ctx, project.ID)
	}), nil
}

// Update is not supported for Projects, as they are immutable.
func (s *REST) Update(ctx kubeapi.Context, obj runtime.Object) (<-chan runtime.Object, error) {
	// TODO handle update of display name, labels, etc.
	return nil, fmt.Errorf("Projects may not be changed.")
}

// Delete asynchronously deletes a Project specified by its id.
func (s *REST) Delete(ctx kubeapi.Context, id string) (<-chan runtime.Object, error) {
	return apiserver.MakeAsync(func() (runtime.Object, error) {
		return &kubeapi.Status{Status: kubeapi.StatusSuccess}, s.registry.DeleteProject(ctx, id)
	}), nil
}
