package clientauthorization

import (
	"fmt"

	kubeapi "github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/apiserver"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/labels"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/runtime"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/util"

	"github.com/openshift/origin/pkg/oauth/api"
)

// REST implements the RESTStorage interface in terms of an Registry.
type REST struct {
	registry Registry
}

// NewStorage returns a new REST.
func NewREST(registry Registry) apiserver.RESTStorage {
	return &REST{registry}
}

// New returns a new ClientAuthorization for use with Create and Update.
func (s *REST) New() runtime.Object {
	return &api.ClientAuthorization{}
}

// Get retrieves an ClientAuthorization by id.
func (s *REST) Get(ctx kubeapi.Context, id string) (runtime.Object, error) {
	authorization, err := s.registry.GetClientAuthorization(id)
	if err != nil {
		return nil, err
	}
	return authorization, nil
}

// List retrieves a list of ClientAuthorizations that match selector.
func (s *REST) List(ctx kubeapi.Context, label, fields labels.Selector) (runtime.Object, error) {
	return s.registry.ListClientAuthorizations(label, labels.Everything())
}

// Create registers the given ClientAuthorization.
func (s *REST) Create(ctx kubeapi.Context, obj runtime.Object) (<-chan runtime.Object, error) {
	authorization, ok := obj.(*api.ClientAuthorization)
	if !ok {
		return nil, fmt.Errorf("not an authorization: %#v", obj)
	}

	if authorization.UserName == "" || authorization.ClientName == "" {
		return nil, fmt.Errorf("invalid authorization")
	}

	authorization.ID = s.registry.ClientAuthorizationID(authorization.UserName, authorization.ClientName)
	authorization.CreationTimestamp = util.Now()

	// if errs := validation.ValidateClientAuthorization(authorization); len(errs) > 0 {
	//  return nil, errors.NewInvalid("clientAuthorization", authorization.Name, errs)
	// }

	return apiserver.MakeAsync(func() (runtime.Object, error) {
		if err := s.registry.CreateClientAuthorization(authorization); err != nil {
			return nil, err
		}
		return s.Get(ctx, authorization.ID)
	}), nil
}

// Update modifies an existing client authorization
func (s *REST) Update(ctx kubeapi.Context, obj runtime.Object) (<-chan runtime.Object, error) {
	return s.Create(ctx, obj)
}

// Delete asynchronously deletes an ClientAuthorization specified by its id.
func (s *REST) Delete(ctx kubeapi.Context, id string) (<-chan runtime.Object, error) {
	return apiserver.MakeAsync(func() (runtime.Object, error) {
		return &kubeapi.Status{Status: kubeapi.StatusSuccess}, s.registry.DeleteClientAuthorization(id)
	}), nil
}
