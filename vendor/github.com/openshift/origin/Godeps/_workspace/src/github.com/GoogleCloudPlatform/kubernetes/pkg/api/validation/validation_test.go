/*
Copyright 2014 Google Inc. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package validation

import (
	"strings"
	"testing"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/api/errors"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/capabilities"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/util"
)

func expectPrefix(t *testing.T, prefix string, errs errors.ErrorList) {
	for i := range errs {
		if !strings.HasPrefix(errs[i].(errors.ValidationError).Field, prefix) {
			t.Errorf("expected prefix '%s' for %v", errs[i])
		}
	}
}

func TestValidateVolumes(t *testing.T) {
	successCase := []api.Volume{
		{Name: "abc"},
		{Name: "123", Source: &api.VolumeSource{HostDir: &api.HostDir{"/mnt/path2"}}},
		{Name: "abc-123", Source: &api.VolumeSource{HostDir: &api.HostDir{"/mnt/path3"}}},
		{Name: "empty", Source: &api.VolumeSource{EmptyDir: &api.EmptyDir{}}},
	}
	names, errs := validateVolumes(successCase)
	if len(errs) != 0 {
		t.Errorf("expected success: %v", errs)
	}
	if len(names) != 4 || !names.HasAll("abc", "123", "abc-123", "empty") {
		t.Errorf("wrong names result: %v", names)
	}

	errorCases := map[string]struct {
		V []api.Volume
		T errors.ValidationErrorType
		F string
	}{
		"zero-length name":     {[]api.Volume{{Name: ""}}, errors.ValidationErrorTypeRequired, "[0].name"},
		"name > 63 characters": {[]api.Volume{{Name: strings.Repeat("a", 64)}}, errors.ValidationErrorTypeInvalid, "[0].name"},
		"name not a DNS label": {[]api.Volume{{Name: "a.b.c"}}, errors.ValidationErrorTypeInvalid, "[0].name"},
		"name not unique":      {[]api.Volume{{Name: "abc"}, {Name: "abc"}}, errors.ValidationErrorTypeDuplicate, "[1].name"},
	}
	for k, v := range errorCases {
		_, errs := validateVolumes(v.V)
		if len(errs) == 0 {
			t.Errorf("expected failure for %s", k)
			continue
		}
		for i := range errs {
			if errs[i].(errors.ValidationError).Type != v.T {
				t.Errorf("%s: expected errors to have type %s: %v", k, v.T, errs[i])
			}
			if errs[i].(errors.ValidationError).Field != v.F {
				t.Errorf("%s: expected errors to have field %s: %v", k, v.F, errs[i])
			}
		}
	}
}

func TestValidatePorts(t *testing.T) {
	successCase := []api.Port{
		{Name: "abc", ContainerPort: 80, HostPort: 80, Protocol: "TCP"},
		{Name: "123", ContainerPort: 81, HostPort: 81},
		{Name: "easy", ContainerPort: 82, Protocol: "TCP"},
		{Name: "as", ContainerPort: 83, Protocol: "UDP"},
		{Name: "do-re-me", ContainerPort: 84},
		{Name: "baby-you-and-me", ContainerPort: 82, Protocol: "tcp"},
		{ContainerPort: 85},
	}
	if errs := validatePorts(successCase); len(errs) != 0 {
		t.Errorf("expected success: %v", errs)
	}

	nonCanonicalCase := []api.Port{
		{ContainerPort: 80},
	}
	if errs := validatePorts(nonCanonicalCase); len(errs) != 0 {
		t.Errorf("expected success: %v", errs)
	}
	if nonCanonicalCase[0].HostPort != 0 || nonCanonicalCase[0].Protocol != "TCP" {
		t.Errorf("expected default values: %+v", nonCanonicalCase[0])
	}

	errorCases := map[string]struct {
		P []api.Port
		T errors.ValidationErrorType
		F string
	}{
		"name > 63 characters": {[]api.Port{{Name: strings.Repeat("a", 64), ContainerPort: 80}}, errors.ValidationErrorTypeInvalid, "[0].name"},
		"name not a DNS label": {[]api.Port{{Name: "a.b.c", ContainerPort: 80}}, errors.ValidationErrorTypeInvalid, "[0].name"},
		"name not unique": {[]api.Port{
			{Name: "abc", ContainerPort: 80},
			{Name: "abc", ContainerPort: 81},
		}, errors.ValidationErrorTypeDuplicate, "[1].name"},
		"zero container port":    {[]api.Port{{ContainerPort: 0}}, errors.ValidationErrorTypeRequired, "[0].containerPort"},
		"invalid container port": {[]api.Port{{ContainerPort: 65536}}, errors.ValidationErrorTypeInvalid, "[0].containerPort"},
		"invalid host port":      {[]api.Port{{ContainerPort: 80, HostPort: 65536}}, errors.ValidationErrorTypeInvalid, "[0].hostPort"},
		"invalid protocol":       {[]api.Port{{ContainerPort: 80, Protocol: "ICMP"}}, errors.ValidationErrorTypeNotSupported, "[0].protocol"},
	}
	for k, v := range errorCases {
		errs := validatePorts(v.P)
		if len(errs) == 0 {
			t.Errorf("expected failure for %s", k)
		}
		for i := range errs {
			if errs[i].(errors.ValidationError).Type != v.T {
				t.Errorf("%s: expected errors to have type %s: %v", k, v.T, errs[i])
			}
			if errs[i].(errors.ValidationError).Field != v.F {
				t.Errorf("%s: expected errors to have field %s: %v", k, v.F, errs[i])
			}
		}
	}
}

func TestValidateEnv(t *testing.T) {
	successCase := []api.EnvVar{
		{Name: "abc", Value: "value"},
		{Name: "ABC", Value: "value"},
		{Name: "AbC_123", Value: "value"},
		{Name: "abc", Value: ""},
	}
	if errs := validateEnv(successCase); len(errs) != 0 {
		t.Errorf("expected success: %v", errs)
	}

	errorCases := map[string][]api.EnvVar{
		"zero-length name":        {{Name: ""}},
		"name not a C identifier": {{Name: "a.b.c"}},
	}
	for k, v := range errorCases {
		if errs := validateEnv(v); len(errs) == 0 {
			t.Errorf("expected failure for %s", k)
		}
	}
}

func TestValidateVolumeMounts(t *testing.T) {
	volumes := util.NewStringSet("abc", "123", "abc-123")

	successCase := []api.VolumeMount{
		{Name: "abc", MountPath: "/foo"},
		{Name: "123", MountPath: "/foo"},
		{Name: "abc-123", MountPath: "/bar"},
	}
	if errs := validateVolumeMounts(successCase, volumes); len(errs) != 0 {
		t.Errorf("expected success: %v", errs)
	}

	errorCases := map[string][]api.VolumeMount{
		"empty name":      {{Name: "", MountPath: "/foo"}},
		"name not found":  {{Name: "", MountPath: "/foo"}},
		"empty mountpath": {{Name: "abc", MountPath: ""}},
	}
	for k, v := range errorCases {
		if errs := validateVolumeMounts(v, volumes); len(errs) == 0 {
			t.Errorf("expected failure for %s", k)
		}
	}
}

func TestValidateContainers(t *testing.T) {
	volumes := util.StringSet{}
	capabilities.SetForTests(capabilities.Capabilities{
		AllowPrivileged: true,
	})

	successCase := []api.Container{
		{Name: "abc", Image: "image"},
		{Name: "123", Image: "image"},
		{Name: "abc-123", Image: "image"},
		{
			Name:  "life-123",
			Image: "image",
			Lifecycle: &api.Lifecycle{
				PreStop: &api.Handler{
					Exec: &api.ExecAction{Command: []string{"ls", "-l"}},
				},
			},
		},
		{Name: "abc-1234", Image: "image", Privileged: true},
	}
	if errs := validateContainers(successCase, volumes); len(errs) != 0 {
		t.Errorf("expected success: %v", errs)
	}

	capabilities.SetForTests(capabilities.Capabilities{
		AllowPrivileged: false,
	})
	errorCases := map[string][]api.Container{
		"zero-length name":     {{Name: "", Image: "image"}},
		"name > 63 characters": {{Name: strings.Repeat("a", 64), Image: "image"}},
		"name not a DNS label": {{Name: "a.b.c", Image: "image"}},
		"name not unique": {
			{Name: "abc", Image: "image"},
			{Name: "abc", Image: "image"},
		},
		"zero-length image": {{Name: "abc", Image: ""}},
		"host port not unique": {
			{Name: "abc", Image: "image", Ports: []api.Port{{ContainerPort: 80, HostPort: 80}}},
			{Name: "def", Image: "image", Ports: []api.Port{{ContainerPort: 81, HostPort: 80}}},
		},
		"invalid env var name": {
			{Name: "abc", Image: "image", Env: []api.EnvVar{{Name: "ev.1"}}},
		},
		"unknown volume name": {
			{Name: "abc", Image: "image", VolumeMounts: []api.VolumeMount{{Name: "anything", MountPath: "/foo"}}},
		},
		"invalid lifecycle, no exec command.": {
			{
				Name:  "life-123",
				Image: "image",
				Lifecycle: &api.Lifecycle{
					PreStop: &api.Handler{
						Exec: &api.ExecAction{},
					},
				},
			},
		},
		"invalid lifecycle, no http path.": {
			{
				Name:  "life-123",
				Image: "image",
				Lifecycle: &api.Lifecycle{
					PreStop: &api.Handler{
						HTTPGet: &api.HTTPGetAction{},
					},
				},
			},
		},
		"invalid lifecycle, no action.": {
			{
				Name:  "life-123",
				Image: "image",
				Lifecycle: &api.Lifecycle{
					PreStop: &api.Handler{},
				},
			},
		},
		"privilege disabled": {
			{Name: "abc", Image: "image", Privileged: true},
		},
	}
	for k, v := range errorCases {
		if errs := validateContainers(v, volumes); len(errs) == 0 {
			t.Errorf("expected failure for %s", k)
		}
	}
}

func TestValidateRestartPolicy(t *testing.T) {
	successCases := []api.RestartPolicy{
		{},
		{Always: &api.RestartPolicyAlways{}},
		{OnFailure: &api.RestartPolicyOnFailure{}},
		{Never: &api.RestartPolicyNever{}},
	}
	for _, policy := range successCases {
		if errs := validateRestartPolicy(&policy); len(errs) != 0 {
			t.Errorf("expected success: %v", errs)
		}
	}

	errorCases := []api.RestartPolicy{
		{Always: &api.RestartPolicyAlways{}, Never: &api.RestartPolicyNever{}},
		{Never: &api.RestartPolicyNever{}, OnFailure: &api.RestartPolicyOnFailure{}},
	}
	for k, policy := range errorCases {
		if errs := validateRestartPolicy(&policy); len(errs) == 0 {
			t.Errorf("expected failure for %s", k)
		}
	}

	noPolicySpecified := api.RestartPolicy{}
	errs := validateRestartPolicy(&noPolicySpecified)
	if len(errs) != 0 {
		t.Errorf("expected success: %v", errs)
	}
	if noPolicySpecified.Always == nil {
		t.Errorf("expected Always policy specified")
	}

}

func TestValidateManifest(t *testing.T) {
	successCases := []api.ContainerManifest{
		{Version: "v1beta1", ID: "abc"},
		{Version: "v1beta2", ID: "123"},
		{Version: "V1BETA1", ID: "abc.123.do-re-mi"},
		{
			Version: "v1beta1",
			ID:      "abc",
			Volumes: []api.Volume{{Name: "vol1", Source: &api.VolumeSource{HostDir: &api.HostDir{"/mnt/vol1"}}},
				{Name: "vol2", Source: &api.VolumeSource{HostDir: &api.HostDir{"/mnt/vol2"}}}},
			Containers: []api.Container{
				{
					Name:       "abc",
					Image:      "image",
					Command:    []string{"foo", "bar"},
					WorkingDir: "/tmp",
					Memory:     1,
					CPU:        1,
					Ports: []api.Port{
						{Name: "p1", ContainerPort: 80, HostPort: 8080},
						{Name: "p2", ContainerPort: 81},
						{ContainerPort: 82},
					},
					Env: []api.EnvVar{
						{Name: "ev1", Value: "val1"},
						{Name: "ev2", Value: "val2"},
						{Name: "EV3", Value: "val3"},
					},
					VolumeMounts: []api.VolumeMount{
						{Name: "vol1", MountPath: "/foo"},
						{Name: "vol1", MountPath: "/bar"},
					},
				},
			},
		},
	}
	for _, manifest := range successCases {
		if errs := ValidateManifest(&manifest); len(errs) != 0 {
			t.Errorf("expected success: %v", errs)
		}
	}

	errorCases := map[string]api.ContainerManifest{
		"empty version":   {Version: "", ID: "abc"},
		"invalid version": {Version: "bogus", ID: "abc"},
		"invalid volume name": {
			Version: "v1beta1",
			ID:      "abc",
			Volumes: []api.Volume{{Name: "vol.1"}},
		},
		"invalid container name": {
			Version:    "v1beta1",
			ID:         "abc",
			Containers: []api.Container{{Name: "ctr.1", Image: "image"}},
		},
	}
	for k, v := range errorCases {
		if errs := ValidateManifest(&v); len(errs) == 0 {
			t.Errorf("expected failure for %s", k)
		}
	}
}

func TestValidatePod(t *testing.T) {
	errs := ValidatePod(&api.Pod{
		JSONBase: api.JSONBase{ID: "foo", Namespace: api.NamespaceDefault},
		Labels: map[string]string{
			"foo": "bar",
		},
		DesiredState: api.PodState{
			Manifest: api.ContainerManifest{
				Version: "v1beta1",
				ID:      "abc",
				RestartPolicy: api.RestartPolicy{
					Always: &api.RestartPolicyAlways{},
				},
			},
		},
	})
	if len(errs) != 0 {
		t.Errorf("Unexpected non-zero error list: %#v", errs)
	}
	errs = ValidatePod(&api.Pod{
		JSONBase: api.JSONBase{ID: "foo", Namespace: api.NamespaceDefault},
		Labels: map[string]string{
			"foo": "bar",
		},
		DesiredState: api.PodState{
			Manifest: api.ContainerManifest{Version: "v1beta1", ID: "abc"},
		},
	})
	if len(errs) != 0 {
		t.Errorf("Unexpected non-zero error list: %#v", errs)
	}

	errs = ValidatePod(&api.Pod{
		JSONBase: api.JSONBase{ID: "foo", Namespace: api.NamespaceDefault},
		Labels: map[string]string{
			"foo": "bar",
		},
		DesiredState: api.PodState{
			Manifest: api.ContainerManifest{
				Version: "v1beta1",
				ID:      "abc",
				RestartPolicy: api.RestartPolicy{Always: &api.RestartPolicyAlways{},
					Never: &api.RestartPolicyNever{}},
			},
		},
	})
	if len(errs) != 1 {
		t.Errorf("Unexpected error list: %#v", errs)
	}
}

func TestValidateService(t *testing.T) {
	testCases := []struct {
		name    string
		svc     api.Service
		numErrs int
	}{
		{
			name: "missing id",
			svc: api.Service{
				JSONBase: api.JSONBase{Namespace: api.NamespaceDefault},
				Port:     8675,
				Selector: map[string]string{"foo": "bar"},
			},
			// Should fail because the ID is missing.
			numErrs: 1,
		},
		{
			name: "missing namespace",
			svc: api.Service{
				JSONBase: api.JSONBase{ID: "foo"},
				Port:     8675,
				Selector: map[string]string{"foo": "bar"},
			},
			// Should fail because the Namespace is missing.
			numErrs: 1,
		},
		{
			name: "invalid id",
			svc: api.Service{
				JSONBase: api.JSONBase{ID: "123abc", Namespace: api.NamespaceDefault},
				Port:     8675,
				Selector: map[string]string{"foo": "bar"},
			},
			// Should fail because the ID is invalid.
			numErrs: 1,
		},
		{
			name: "missing port",
			svc: api.Service{
				JSONBase: api.JSONBase{ID: "abc123", Namespace: api.NamespaceDefault},
				Selector: map[string]string{"foo": "bar"},
			},
			// Should fail because the port number is missing/invalid.
			numErrs: 1,
		},
		{
			name: "invalid port",
			svc: api.Service{
				JSONBase: api.JSONBase{ID: "abc123", Namespace: api.NamespaceDefault},
				Port:     65536,
				Selector: map[string]string{"foo": "bar"},
			},
			// Should fail because the port number is invalid.
			numErrs: 1,
		},
		{
			name: "invalid protocol",
			svc: api.Service{
				JSONBase: api.JSONBase{ID: "abc123", Namespace: api.NamespaceDefault},
				Port:     8675,
				Protocol: "INVALID",
				Selector: map[string]string{"foo": "bar"},
			},
			// Should fail because the protocol is invalid.
			numErrs: 1,
		},
		{
			name: "missing selector",
			svc: api.Service{
				JSONBase: api.JSONBase{ID: "foo", Namespace: api.NamespaceDefault},
				Port:     8675,
			},
			// Should fail because the selector is missing.
			numErrs: 1,
		},
		{
			name: "valid 1",
			svc: api.Service{
				JSONBase: api.JSONBase{ID: "abc123", Namespace: api.NamespaceDefault},
				Port:     1,
				Protocol: "TCP",
				Selector: map[string]string{"foo": "bar"},
			},
			numErrs: 0,
		},
		{
			name: "valid 2",
			svc: api.Service{
				JSONBase: api.JSONBase{ID: "abc123", Namespace: api.NamespaceDefault},
				Port:     65535,
				Protocol: "UDP",
				Selector: map[string]string{"foo": "bar"},
			},
			numErrs: 0,
		},
		{
			name: "valid 3",
			svc: api.Service{
				JSONBase: api.JSONBase{ID: "abc123", Namespace: api.NamespaceDefault},
				Port:     80,
				Selector: map[string]string{"foo": "bar"},
			},
			numErrs: 0,
		},
	}

	for _, tc := range testCases {
		errs := ValidateService(&tc.svc)
		if len(errs) != tc.numErrs {
			t.Errorf("Unexpected error list for case %q: %+v", tc.name, errs)
		}
	}

	svc := api.Service{
		Port:     6502,
		JSONBase: api.JSONBase{ID: "foo", Namespace: api.NamespaceDefault},
		Selector: map[string]string{"foo": "bar"},
	}
	errs := ValidateService(&svc)
	if len(errs) != 0 {
		t.Errorf("Unexpected non-zero error list: %#v", errs)
	}
	if svc.Protocol != "TCP" {
		t.Errorf("Expected default protocol of 'TCP': %#v", errs)
	}
}

func TestValidateReplicationController(t *testing.T) {
	validSelector := map[string]string{"a": "b"}
	validPodTemplate := api.PodTemplate{
		DesiredState: api.PodState{
			Manifest: api.ContainerManifest{
				Version: "v1beta1",
			},
		},
		Labels: validSelector,
	}

	successCases := []api.ReplicationController{
		{
			JSONBase: api.JSONBase{ID: "abc", Namespace: api.NamespaceDefault},
			DesiredState: api.ReplicationControllerState{
				ReplicaSelector: validSelector,
				PodTemplate:     validPodTemplate,
			},
		},
		{
			JSONBase: api.JSONBase{ID: "abc-123", Namespace: api.NamespaceDefault},
			DesiredState: api.ReplicationControllerState{
				ReplicaSelector: validSelector,
				PodTemplate:     validPodTemplate,
			},
		},
	}
	for _, successCase := range successCases {
		if errs := ValidateReplicationController(&successCase); len(errs) != 0 {
			t.Errorf("expected success: %v", errs)
		}
	}

	errorCases := map[string]api.ReplicationController{
		"zero-length ID": {
			JSONBase: api.JSONBase{ID: "", Namespace: api.NamespaceDefault},
			DesiredState: api.ReplicationControllerState{
				ReplicaSelector: validSelector,
				PodTemplate:     validPodTemplate,
			},
		},
		"missing-namespace": {
			JSONBase: api.JSONBase{ID: "abc-123"},
			DesiredState: api.ReplicationControllerState{
				ReplicaSelector: validSelector,
				PodTemplate:     validPodTemplate,
			},
		},
		"empty selector": {
			JSONBase: api.JSONBase{ID: "abc", Namespace: api.NamespaceDefault},
			DesiredState: api.ReplicationControllerState{
				PodTemplate: validPodTemplate,
			},
		},
		"selector_doesnt_match": {
			JSONBase: api.JSONBase{ID: "abc", Namespace: api.NamespaceDefault},
			DesiredState: api.ReplicationControllerState{
				ReplicaSelector: map[string]string{"foo": "bar"},
				PodTemplate:     validPodTemplate,
			},
		},
		"invalid manifest": {
			JSONBase: api.JSONBase{ID: "abc", Namespace: api.NamespaceDefault},
			DesiredState: api.ReplicationControllerState{
				ReplicaSelector: validSelector,
			},
		},
		"negative_replicas": {
			JSONBase: api.JSONBase{ID: "abc", Namespace: api.NamespaceDefault},
			DesiredState: api.ReplicationControllerState{
				Replicas:        -1,
				ReplicaSelector: validSelector,
			},
		},
	}
	for k, v := range errorCases {
		errs := ValidateReplicationController(&v)
		if len(errs) == 0 {
			t.Errorf("expected failure for %s", k)
		}
		for i := range errs {
			field := errs[i].(errors.ValidationError).Field
			if !strings.HasPrefix(field, "desiredState.podTemplate.") &&
				field != "id" &&
				field != "namespace" &&
				field != "desiredState.replicaSelector" &&
				field != "desiredState.replicas" {
				t.Errorf("%s: missing prefix for: %v", k, errs[i])
			}
		}
	}
}
