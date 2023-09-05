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

	"github.com/sirupsen/logrus"
	"github.com/uwu-tools/peribolos/third_party/k8s.io/test-infra/prow/config/org"
	"github.com/uwu-tools/peribolos/third_party/k8s.io/test-infra/prow/github"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/uwu-tools/peribolos/options/root"
)

func Configure(opt root.Options, client github.Client, orgName string, orgConfig org.Config) error {
	// Ensure that metadata is configured correctly.
	if !opt.FixOrg {
		logrus.Infof("Skipping org metadata configuration")
	} else if err := configureOrgMeta(client, orgName, orgConfig.Metadata); err != nil {
		return err
	}

	invitees, err := orgInvitations(opt, client, orgName)
	if err != nil {
		return fmt.Errorf("failed to list %s invitations: %w", orgName, err)
	}

	// Invite/remove/update members to the org.
	if !opt.FixOrgMembers {
		logrus.Infof("Skipping org member configuration")
	} else if err := configureOrgMembers(opt, client, orgName, orgConfig, invitees); err != nil {
		return fmt.Errorf("failed to configure %s members: %w", orgName, err)
	}

	// Create repositories in the org
	if !opt.FixRepos {
		logrus.Info("Skipping org repositories configuration")
	} else if err := configureRepos(opt, client, orgName, orgConfig); err != nil {
		return fmt.Errorf("failed to configure %s repos: %w", orgName, err)
	}

	if !opt.FixTeams {
		logrus.Infof("Skipping team and team member configuration")
		return nil
	}

	// Find the id and current state of each declared team (create/delete as necessary)
	githubTeams, err := configureTeams(client, orgName, orgConfig, opt.MaxDelta, opt.IgnoreSecretTeams)
	if err != nil {
		return fmt.Errorf("failed to configure %s teams: %w", orgName, err)
	}

	for name, team := range orgConfig.Teams {
		err := configureTeamAndMembers(opt, client, githubTeams, name, orgName, team, nil)
		if err != nil {
			return fmt.Errorf("failed to configure %s teams: %w", orgName, err)
		}

		if !opt.FixTeamRepos {
			logrus.Infof("Skipping team repo permissions configuration")
			continue
		}
		if err := configureTeamRepos(opt, client, githubTeams, name, orgName, team); err != nil {
			return fmt.Errorf("failed to configure %s team %s repos: %w", orgName, name, err)
		}
	}
	return nil
}

type orgMetadataClient interface {
	GetOrg(name string) (*github.Organization, error)
	EditOrg(name string, org github.Organization) (*github.Organization, error)
}

// configureOrgMeta will update github to have the non-nil wanted metadata values.
func configureOrgMeta(client orgMetadataClient, orgName string, want org.Metadata) error {
	cur, err := client.GetOrg(orgName)
	if err != nil {
		return fmt.Errorf("failed to get %s metadata: %w", orgName, err)
	}
	change := false
	change = updateString(&cur.BillingEmail, want.BillingEmail) || change
	change = updateString(&cur.Company, want.Company) || change
	change = updateString(&cur.Email, want.Email) || change
	change = updateString(&cur.Name, want.Name) || change
	change = updateString(&cur.Description, want.Description) || change
	change = updateString(&cur.Location, want.Location) || change
	if want.DefaultRepositoryPermission != nil {
		w := string(*want.DefaultRepositoryPermission)
		change = updateString(&cur.DefaultRepositoryPermission, &w) || change
	}
	change = updateBool(&cur.HasOrganizationProjects, want.HasOrganizationProjects) || change
	change = updateBool(&cur.HasRepositoryProjects, want.HasRepositoryProjects) || change
	change = updateBool(&cur.MembersCanCreateRepositories, want.MembersCanCreateRepositories) || change
	if change {
		if _, err := client.EditOrg(orgName, *cur); err != nil {
			return fmt.Errorf("failed to edit %s metadata: %w", orgName, err)
		}
	}
	return nil
}

// Helpers

// updateString will return true and set have to want iff they are set and different.
func updateString(have, want *string) bool {
	switch {
	case have == nil:
		panic("have must be non-nil")
	case want == nil:
		return false // do not care what we have
	case *have == *want:
		return false // already have it
	}
	*have = *want // update value
	return true
}

// updateBool will return true and set have to want iff they are set and different.
func updateBool(have, want *bool) bool {
	switch {
	case have == nil:
		panic("have must not be nil")
	case want == nil:
		return false // do not care what we have
	case *have == *want:
		return false // already have it
	}
	*have = *want // update value
	return true
}

func normalize(s sets.Set[string]) sets.Set[string] {
	out := sets.Set[string]{}
	for i := range s {
		out.Insert(github.NormLogin(i))
	}
	return out
}
