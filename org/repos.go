/*
Copyright RelEngFam Authors.

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

package org

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/uwu-tools/peribolos/third_party/k8s.io/test-infra/prow/config/org"
	"github.com/uwu-tools/peribolos/third_party/k8s.io/test-infra/prow/github"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"

	"github.com/uwu-tools/peribolos/options/root"
)

type repoClient interface {
	GetRepo(orgName, repo string) (github.FullRepo, error)
	GetRepos(orgName string, isUser bool) ([]github.Repo, error)
	CreateRepo(owner string, isUser bool, repo github.RepoCreateRequest) (*github.FullRepo, error)
	UpdateRepo(owner, name string, repo github.RepoUpdateRequest) (*github.FullRepo, error)
}

func configureRepos(opt root.Options, client repoClient, orgName string, orgConfig org.Config) error {
	if err := validateRepos(orgConfig.Repos); err != nil {
		return err
	}

	repoList, err := client.GetRepos(orgName, false)
	if err != nil {
		return fmt.Errorf("failed to get repos: %w", err)
	}
	logrus.Debugf("Found %d repositories", len(repoList))
	byName := make(map[string]github.Repo, len(repoList))
	for _, repo := range repoList {
		byName[strings.ToLower(repo.Name)] = repo
	}

	var allErrors []error

	for wantName, wantRepo := range orgConfig.Repos {
		repoLogger := logrus.WithField("repo", wantName)
		pastErrors := len(allErrors)
		var existing *github.FullRepo = nil
		for _, possibleName := range append([]string{wantName}, wantRepo.Previously...) {
			if repo, exists := byName[strings.ToLower(possibleName)]; exists {
				switch {
				case existing == nil:
					if full, err := client.GetRepo(orgName, repo.Name); err != nil {
						repoLogger.WithError(err).Error("failed to get repository data")
						allErrors = append(allErrors, err)
					} else {
						existing = &full
					}
				case existing.Name != repo.Name:
					err := fmt.Errorf("different repos already exist for current and previous names: %s and %s", existing.Name, repo.Name)
					allErrors = append(allErrors, err)
				}
			}
		}

		if len(allErrors) > pastErrors {
			continue
		}

		if existing == nil {
			if wantRepo.Archived != nil && *wantRepo.Archived {
				repoLogger.Error("repo does not exist but is configured as archived: not creating")
				allErrors = append(allErrors, fmt.Errorf("nonexistent repo configured as archived: %s", wantName))
				continue
			}
			repoLogger.Info("repo does not exist, creating")
			created, err := client.CreateRepo(orgName, false, newRepoCreateRequest(wantName, wantRepo))
			if err != nil {
				repoLogger.WithError(err).Error("failed to create repository")
				allErrors = append(allErrors, err)
			} else {
				existing = created
			}
		}

		if existing != nil {
			if existing.Archived {
				if wantRepo.Archived != nil && *wantRepo.Archived {
					repoLogger.Infof("repo %q is archived, skipping changes", wantName)
					continue
				}
			}
			repoLogger.Info("repo exists, considering an update")
			delta := newRepoUpdateRequest(*existing, wantName, wantRepo)
			if deltaErrors := sanitizeRepoDelta(opt, &delta); len(deltaErrors) > 0 {
				for _, err := range deltaErrors {
					repoLogger.WithError(err).Error("requested repo change is not allowed, removing from delta")
				}
				allErrors = append(allErrors, deltaErrors...)
			}
			if delta.Defined() {
				repoLogger.Info("repo exists and differs from desired state, updating")
				if _, err := client.UpdateRepo(orgName, existing.Name, delta); err != nil {
					repoLogger.WithError(err).Error("failed to update repository")
					allErrors = append(allErrors, err)
				}
			}
		}
	}

	return utilerrors.NewAggregate(allErrors)
}

func validateRepos(repos map[string]org.Repo) error {
	seen := map[string]string{}
	var dups []string

	for wantName, repo := range repos {
		toCheck := append([]string{wantName}, repo.Previously...)
		for _, name := range toCheck {
			normName := strings.ToLower(name)
			if seenName, have := seen[normName]; have {
				dups = append(dups, fmt.Sprintf("%s/%s", seenName, name))
			}
		}
		for _, name := range toCheck {
			normName := strings.ToLower(name)
			seen[normName] = name
		}
	}

	if len(dups) > 0 {
		return fmt.Errorf("found duplicate repo names (GitHub repo names are case-insensitive): %s", strings.Join(dups, ", "))
	}

	return nil
}

func newRepoCreateRequest(name string, definition org.Repo) github.RepoCreateRequest {
	repoCreate := github.RepoCreateRequest{
		RepoRequest: github.RepoRequest{
			Name:                     &name,
			Description:              definition.Description,
			Homepage:                 definition.HomePage,
			Private:                  definition.Private,
			HasIssues:                definition.HasIssues,
			HasProjects:              definition.HasProjects,
			HasWiki:                  definition.HasWiki,
			AllowSquashMerge:         definition.AllowSquashMerge,
			AllowMergeCommit:         definition.AllowMergeCommit,
			AllowRebaseMerge:         definition.AllowRebaseMerge,
			SquashMergeCommitTitle:   definition.SquashMergeCommitTitle,
			SquashMergeCommitMessage: definition.SquashMergeCommitMessage,
		},
	}

	if definition.OnCreate != nil {
		repoCreate.AutoInit = definition.OnCreate.AutoInit
		repoCreate.GitignoreTemplate = definition.OnCreate.GitignoreTemplate
		repoCreate.LicenseTemplate = definition.OnCreate.LicenseTemplate
	}

	return repoCreate
}

// newRepoUpdateRequest creates a minimal github.RepoUpdateRequest instance
// needed to update the current repo into the target state.
func newRepoUpdateRequest(current github.FullRepo, name string, repo org.Repo) github.RepoUpdateRequest {
	setString := func(current string, want *string) *string {
		if want != nil && *want != current {
			return want
		}
		return nil
	}
	setBool := func(current bool, want *bool) *bool {
		if want != nil && *want != current {
			return want
		}
		return nil
	}
	repoUpdate := github.RepoUpdateRequest{
		RepoRequest: github.RepoRequest{
			Name:                     setString(current.Name, &name),
			Description:              setString(current.Description, repo.Description),
			Homepage:                 setString(current.Homepage, repo.HomePage),
			Private:                  setBool(current.Private, repo.Private),
			HasIssues:                setBool(current.HasIssues, repo.HasIssues),
			HasProjects:              setBool(current.HasProjects, repo.HasProjects),
			HasWiki:                  setBool(current.HasWiki, repo.HasWiki),
			AllowSquashMerge:         setBool(current.AllowSquashMerge, repo.AllowSquashMerge),
			AllowMergeCommit:         setBool(current.AllowMergeCommit, repo.AllowMergeCommit),
			AllowRebaseMerge:         setBool(current.AllowRebaseMerge, repo.AllowRebaseMerge),
			SquashMergeCommitTitle:   setString(current.SquashMergeCommitTitle, repo.SquashMergeCommitTitle),
			SquashMergeCommitMessage: setString(current.SquashMergeCommitMessage, repo.SquashMergeCommitMessage),
		},
		DefaultBranch: setString(current.DefaultBranch, repo.DefaultBranch),
		Archived:      setBool(current.Archived, repo.Archived),
	}

	return repoUpdate
}

func sanitizeRepoDelta(opt root.Options, delta *github.RepoUpdateRequest) []error {
	var errs []error
	if delta.Archived != nil && !*delta.Archived {
		delta.Archived = nil
		errs = append(errs, fmt.Errorf("asked to unarchive an archived repo, unsupported by GH API"))
	}
	if delta.Archived != nil && *delta.Archived && !opt.AllowRepoArchival {
		delta.Archived = nil
		errs = append(errs, fmt.Errorf("asked to archive a repo but this is not allowed by default (see --allow-repo-archival)"))
	}
	if delta.Private != nil && !(*delta.Private || opt.AllowRepoPublish) {
		delta.Private = nil
		errs = append(errs, fmt.Errorf("asked to publish a private repo but this is not allowed by default (see --allow-repo-publish)"))
	}

	return errs
}
