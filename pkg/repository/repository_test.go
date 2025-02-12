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

package repository

import (
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/stretchr/testify/assert"
)

type fakeRepositoryGetter struct {
	repository *fakeRepository
	err        error
}

func (frg fakeRepositoryGetter) get(_ string) (gitRepositoryInterface, error) {
	return frg.repository, frg.err
}

type fakeRepository struct {
	worktree *fakeWorktree
	head     *plumbing.Reference
	err      error
}

func (fr fakeRepository) Worktree() (gitWorktreeInterface, error) {
	return fr.worktree, fr.err
}

func (fr fakeRepository) Head() (*plumbing.Reference, error) {
	return fr.head, fr.err
}

type fakeWorktree struct {
	status *fakeStatus
	err    error
}

func (fw fakeWorktree) Status() (gitStatusInterface, error) {
	return fw.status, fw.err
}

type fakeStatus struct {
	isClean bool
}

func (fs fakeStatus) IsClean() bool {
	return fs.isClean
}

func TestIsClean(t *testing.T) {
	type config struct {
		repositoryGetter fakeRepositoryGetter
	}
	type expected struct {
		isClean bool
		err     error
	}
	var tests = []struct {
		name     string
		config   config
		expected expected
	}{
		{
			name: "dir is not a repository",
			config: config{
				repositoryGetter: fakeRepositoryGetter{
					repository: nil,
					err:        git.ErrRepositoryNotExists,
				},
			},
			expected: expected{
				isClean: false,
				err:     git.ErrRepositoryNotExists,
			},
		},
		{
			name: "repository could not access worktree",
			config: config{
				repositoryGetter: fakeRepositoryGetter{
					repository: &fakeRepository{
						worktree: nil,
						err:      assert.AnError,
					},
				},
			},
			expected: expected{
				isClean: false,
				err:     assert.AnError,
			},
		},
		{
			name: "worktree could not access status",
			config: config{
				repositoryGetter: fakeRepositoryGetter{
					repository: &fakeRepository{
						worktree: &fakeWorktree{
							status: nil,
							err:    assert.AnError,
						},
						err: nil,
					},
				},
			},
			expected: expected{
				isClean: false,
				err:     assert.AnError,
			},
		},
		{
			name: "repository is not clean",
			config: config{
				repositoryGetter: fakeRepositoryGetter{
					repository: &fakeRepository{
						worktree: &fakeWorktree{
							status: &fakeStatus{
								isClean: false,
							},
							err: nil,
						},
						err: nil,
					},
				},
			},
			expected: expected{
				isClean: false,
				err:     nil,
			},
		},
		{
			name: "repository is clean",
			config: config{
				repositoryGetter: fakeRepositoryGetter{
					repository: &fakeRepository{
						worktree: &fakeWorktree{
							status: &fakeStatus{
								isClean: true,
							},
							err: nil,
						},
						err: nil,
					},
				},
			},
			expected: expected{
				isClean: true,
				err:     nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := Repository{
				repositoryGetter: tt.config.repositoryGetter,
			}
			isClean, err := repo.IsClean()
			assert.ErrorIs(t, err, tt.expected.err)
			assert.Equal(t, tt.expected.isClean, isClean)
		})
	}
}

func TestGetSHA(t *testing.T) {
	type config struct {
		repositoryGetter fakeRepositoryGetter
	}
	type expected struct {
		sha string
		err error
	}
	var tests = []struct {
		name     string
		config   config
		expected expected
	}{
		{
			name: "get sha without any problem",
			config: config{
				repositoryGetter: fakeRepositoryGetter{
					repository: &fakeRepository{
						head: plumbing.NewHashReference("test", plumbing.NewHash("test")),
					},
				},
			},
			expected: expected{
				sha: plumbing.NewHash("test").String(),
				err: nil,
			},
		},
		{
			name: "error getting repository",
			config: config{
				repositoryGetter: fakeRepositoryGetter{
					repository: nil,
					err:        assert.AnError,
				},
			},
			expected: expected{
				sha: "",
				err: assert.AnError,
			},
		},
		{
			name: "error getting Head",
			config: config{
				repositoryGetter: fakeRepositoryGetter{
					repository: &fakeRepository{
						head: nil,
						err:  assert.AnError,
					},
				},
			},
			expected: expected{
				sha: "",
				err: assert.AnError,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := Repository{
				repositoryGetter: tt.config.repositoryGetter,
			}
			sha, err := repo.GetSHA()
			assert.ErrorIs(t, err, tt.expected.err)
			assert.Equal(t, tt.expected.sha, sha)
		})
	}
}
