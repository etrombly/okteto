// Copyright 2023 The Okteto Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package okteto

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/okteto/okteto/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestDeployPreview(t *testing.T) {
	type input struct {
		client    *fakeGraphQLClient
		name      string
		variables []types.Variable
	}
	type expected struct {
		response *types.PreviewResponse
		err      error
	}
	testCases := []struct {
		name     string
		input    input
		expected expected
	}{
		{
			name: "namespace validator length exceeds",
			input: input{
				name: strings.Repeat("a", 100),
			},
			expected: expected{
				response: nil,
				err: namespaceValidationError{
					object: previewEnvObject,
				},
			},
		},
		{
			name: "namespace validator does not match regexp",
			input: input{
				name: "-",
			},
			expected: expected{
				response: nil,
				err: namespaceValidationError{
					object: previewEnvObject,
				},
			},
		},
		{
			name: "with variables - error",
			input: input{
				client: &fakeGraphQLClient{
					err: assert.AnError,
				},
				name: "test",
				variables: []types.Variable{
					{
						Name:  "OKTETO_ORIGIN",
						Value: "VALUE",
					},
				},
			},
			expected: expected{
				response: nil,
				err:      assert.AnError,
			},
		},
		{
			name: "with variables - no error",
			input: input{
				client: &fakeGraphQLClient{
					mutationResult: &deployPreviewMutation{
						Response: deployPreviewResponse{
							Action: actionStruct{
								Id:     "test",
								Name:   "test",
								Status: ProgressingStatus,
							},
							Preview: previewIDStruct{
								Id: "test",
							},
						},
					},
					err: nil,
				},
				name: "test",
				variables: []types.Variable{
					{
						Name:  "KEY",
						Value: "VALUE",
					},
				},
			},
			expected: expected{
				response: &types.PreviewResponse{
					Action: &types.Action{
						ID:     "test",
						Name:   "test",
						Status: progressingStatus,
					},
					Preview: &types.Preview{
						ID: "test",
					},
				},
				err: nil,
			},
		},
		{
			name: "without variables - error",
			input: input{
				client: &fakeGraphQLClient{
					err: assert.AnError,
				},
				name:      "test",
				variables: []types.Variable{},
			},
			expected: expected{
				response: nil,
				err:      assert.AnError,
			},
		},
		{
			name: "without variables - no error",
			input: input{
				client: &fakeGraphQLClient{
					mutationResult: &deployPreviewMutation{
						Response: deployPreviewResponse{
							Action: actionStruct{
								Id:     "test",
								Name:   "test",
								Status: ProgressingStatus,
							},
							Preview: previewIDStruct{
								Id: "test",
							},
						},
					},
					err: nil,
				},
				name:      "test",
				variables: []types.Variable{},
			},
			expected: expected{
				response: &types.PreviewResponse{
					Action: &types.Action{
						ID:     "test",
						Name:   "test",
						Status: progressingStatus,
					},
					Preview: &types.Preview{
						ID: "test",
					},
				},
				err: nil,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pc := previewClient{
				client:             tc.input.client,
				namespaceValidator: newNamespaceValidator(),
			}
			response, err := pc.DeployPreview(context.Background(), tc.input.name, "", "", "", "", "", tc.input.variables)
			assert.ErrorIs(t, err, tc.expected.err)
			assert.Equal(t, tc.expected.response, response)
		})
	}
}

func TestDestroyPreview(t *testing.T) {
	type input struct {
		client *fakeGraphQLClient
		name   string
	}
	type expected struct {
		err error
	}
	testCases := []struct {
		name     string
		input    input
		expected expected
	}{
		{
			name: "no error",
			input: input{
				client: &fakeGraphQLClient{
					mutationResult: &destroyPreviewMutation{
						Response: previewIDStruct{
							Id: "test",
						},
					},
					err: nil,
				},
				name: "test",
			},
			expected: expected{
				err: nil,
			},
		},
		{
			name: "error",
			input: input{
				client: &fakeGraphQLClient{
					mutationResult: nil,
					err:            assert.AnError,
				},
				name: "test",
			},
			expected: expected{
				err: assert.AnError,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pc := previewClient{
				client: tc.input.client,
			}
			err := pc.Destroy(context.Background(), tc.input.name)
			assert.ErrorIs(t, err, tc.expected.err)
		})
	}
}

func TestListPreview(t *testing.T) {
	type input struct {
		client *fakeGraphQLClient
	}
	type expected struct {
		response []types.Preview
		err      error
	}
	testCases := []struct {
		name     string
		input    input
		expected expected
	}{
		{
			name: "no error",
			input: input{
				client: &fakeGraphQLClient{
					queryResult: &listPreviewQuery{
						Response: []previewEnv{
							{
								Id:       "test",
								Sleeping: false,
								Scope:    "test",
							},
						},
					},
					err: nil,
				},
			},
			expected: expected{
				response: []types.Preview{
					{
						ID:       "test",
						Sleeping: false,
						Scope:    "test",
					},
				},
				err: nil,
			},
		},
		{
			name: "error",
			input: input{
				client: &fakeGraphQLClient{
					queryResult: nil,
					err:         assert.AnError,
				},
			},
			expected: expected{
				err: assert.AnError,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pc := previewClient{
				client: tc.input.client,
			}
			response, err := pc.List(context.Background())
			assert.ErrorIs(t, err, tc.expected.err)
			assert.Equal(t, tc.expected.response, response)
		})
	}
}

func TestListEndpoints(t *testing.T) {
	type input struct {
		client *fakeGraphQLClient
		name   string
	}
	type expected struct {
		response []types.Endpoint
		err      error
	}
	testCases := []struct {
		name     string
		input    input
		expected expected
	}{
		{
			name: "no error",
			input: input{
				client: &fakeGraphQLClient{
					queryResult: &listPreviewEndpoints{
						Response: previewEndpoints{
							Deployments: []deploymentEndpoint{
								{
									Endpoints: []endpointURL{
										{
											Url: "https://test.test",
										},
									},
								},
							},
							Statefulsets: []statefulsetEdnpoint{
								{
									Endpoints: []endpointURL{
										{
											Url: "https://test.test",
										},
									},
								},
							},
						},
					},
					err: nil,
				},
				name: "test",
			},
			expected: expected{
				response: []types.Endpoint{
					{
						URL: "https://test.test",
					},
					{
						URL: "https://test.test",
					},
				},
				err: nil,
			},
		},
		{
			name: "error",
			input: input{
				client: &fakeGraphQLClient{
					queryResult: nil,
					err:         assert.AnError,
				},
				name: "test",
			},
			expected: expected{
				err: assert.AnError,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pc := previewClient{
				client: tc.input.client,
			}
			response, err := pc.ListEndpoints(context.Background(), tc.input.name)
			assert.ErrorIs(t, err, tc.expected.err)
			assert.Equal(t, tc.expected.response, response)
		})
	}
}

func TestGetResourcesStatus(t *testing.T) {
	type input struct {
		client     *fakeGraphQLClient
		namespace  string
		devenvName string
	}
	type expected struct {
		response map[string]string
		err      error
	}
	testCases := []struct {
		name     string
		input    input
		expected expected
	}{
		{
			name: "error",
			input: input{
				client: &fakeGraphQLClient{
					queryResult: nil,
					err:         assert.AnError,
				},
			},
			expected: expected{
				err: assert.AnError,
			},
		},
		{
			name: "no error - empty devenv - return all the resources",
			input: input{
				client: &fakeGraphQLClient{
					queryResult: &getPreviewResources{
						Response: previewResourcesStatus{
							Deployments: []resourceInfo{
								{
									ID:         "test",
									Name:       "error",
									Status:     "error",
									DeployedBy: "test",
								},
								{
									ID:         "test",
									Name:       "queued",
									Status:     "queued",
									DeployedBy: "test",
								},
								{
									ID:         "test",
									Name:       "progressing",
									Status:     "progressing",
									DeployedBy: "test",
								},
							},
							Statefulsets: []resourceInfo{
								{
									ID:         "test",
									Name:       "error",
									Status:     "error",
									DeployedBy: "test",
								},
								{
									ID:         "test",
									Name:       "queued",
									Status:     "queued",
									DeployedBy: "test",
								},
								{
									ID:         "test",
									Name:       "progressing",
									Status:     "progressing",
									DeployedBy: "test",
								},
							},
							Jobs: []resourceInfo{
								{
									ID:         "test",
									Name:       "error",
									Status:     "error",
									DeployedBy: "test",
								},
								{
									ID:         "test",
									Name:       "queued",
									Status:     "queued",
									DeployedBy: "test",
								},
								{
									ID:         "test",
									Name:       "progressing",
									Status:     "progressing",
									DeployedBy: "test",
								},
							},
							Cronjobs: []resourceInfo{
								{
									ID:         "test",
									Name:       "error",
									Status:     "error",
									DeployedBy: "test",
								},
								{
									ID:         "test",
									Name:       "queued",
									Status:     "queued",
									DeployedBy: "test",
								},
								{
									ID:         "test",
									Name:       "progressing",
									Status:     "progressing",
									DeployedBy: "test",
								},
							},
						},
					},
					err: nil,
				},
				namespace:  "test",
				devenvName: "test",
			},
			expected: expected{
				response: map[string]string{
					"deployment/error":        "error",
					"deployment/queued":       "queued",
					"deployment/progressing":  "progressing",
					"statefulset/error":       "error",
					"statefulset/queued":      "queued",
					"statefulset/progressing": "progressing",
					"job/error":               "error",
					"job/queued":              "queued",
					"job/progressing":         "progressing",
					"cronjob/error":           "error",
					"cronjob/queued":          "queued",
					"cronjob/progressing":     "progressing",
				},
				err: nil,
			},
		},
		{
			name: "no error - devenv only deploy errors",
			input: input{
				client: &fakeGraphQLClient{
					queryResult: &getPreviewResources{
						Response: previewResourcesStatus{
							Deployments: []resourceInfo{
								{
									ID:         "test",
									Name:       "error",
									Status:     "error",
									DeployedBy: "1",
								},
								{
									ID:         "test",
									Name:       "queued",
									Status:     "queued",
									DeployedBy: "2",
								},
								{
									ID:         "test",
									Name:       "progressing",
									Status:     "progressing",
									DeployedBy: "3",
								},
							},
							Statefulsets: []resourceInfo{
								{
									ID:         "test",
									Name:       "error",
									Status:     "error",
									DeployedBy: "1",
								},
								{
									ID:         "test",
									Name:       "queued",
									Status:     "queued",
									DeployedBy: "2",
								},
								{
									ID:         "test",
									Name:       "progressing",
									Status:     "progressing",
									DeployedBy: "3",
								},
							},
							Jobs: []resourceInfo{
								{
									ID:         "test",
									Name:       "error",
									Status:     "error",
									DeployedBy: "1",
								},
								{
									ID:         "test",
									Name:       "queued",
									Status:     "queued",
									DeployedBy: "2",
								},
								{
									ID:         "test",
									Name:       "progressing",
									Status:     "progressing",
									DeployedBy: "3",
								},
							},
							Cronjobs: []resourceInfo{
								{
									ID:         "test",
									Name:       "error",
									Status:     "error",
									DeployedBy: "1",
								},
								{
									ID:         "test",
									Name:       "queued",
									Status:     "queued",
									DeployedBy: "2",
								},
								{
									ID:         "test",
									Name:       "progressing",
									Status:     "progressing",
									DeployedBy: "3",
								},
							},
						},
					},
					err: nil,
				},
				namespace:  "test",
				devenvName: "1",
			},
			expected: expected{
				response: map[string]string{
					"deployment/error":  "error",
					"statefulset/error": "error",
					"job/error":         "error",
					"cronjob/error":     "error",
				},
				err: nil,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pc := previewClient{
				client: tc.input.client,
			}
			response, err := pc.GetResourcesStatus(context.Background(), tc.input.namespace, tc.input.devenvName)
			assert.ErrorIs(t, err, tc.expected.err)
			assert.Equal(t, tc.expected.response, response)
		})
	}
}

func TestTranslatePreviewErr(t *testing.T) {
	type input struct {
		err  error
		name string
	}
	type expected struct {
		err error
	}
	testCases := []struct {
		name     string
		input    input
		expected expected
	}{
		{
			name: "another error",
			input: input{
				err:  assert.AnError,
				name: "test",
			},
			expected: expected{
				err: assert.AnError,
			},
		},
		{
			name: "conflict",
			input: input{
				err:  errors.New("conflict"),
				name: "test",
			},
			expected: expected{
				err: previewConflictErr{
					name: "test",
				},
			},
		},
		{
			name: "operation-not-permitted",
			input: input{
				err:  errors.New("operation-not-permitted"),
				name: "test",
			},
			expected: expected{
				err: ErrUnauthorizedGlobalCreation,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pc := previewClient{}
			err := pc.translateErr(tc.input.err, tc.input.name)
			assert.ErrorIs(t, err, tc.expected.err)
		})
	}
}
