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
	"k8s.io/test-infra/prow/config/org"
	"k8s.io/test-infra/prow/github"
)

type dumpClient interface {
	GetOrg(name string) (*github.Organization, error)
	ListOrgMembers(org, role string) ([]github.TeamMember, error)
	ListTeams(org string) ([]github.Team, error)
	ListTeamMembersBySlug(org, teamSlug, role string) ([]github.TeamMember, error)
	ListTeamReposBySlug(org, teamSlug string) ([]github.Repo, error)
	GetRepo(owner, name string) (github.FullRepo, error)
	GetRepos(org string, isUser bool) ([]github.Repo, error)
	BotUser() (*github.UserData, error)
}

func Dump(client dumpClient, orgName string, ignoreSecretTeams bool, appID string) (*org.Config, error) {
	out := org.Config{}
	meta, err := client.GetOrg(orgName)
	if err != nil {
		return nil, fmt.Errorf("failed to get org: %w", err)
	}
	out.Metadata.BillingEmail = &meta.BillingEmail
	out.Metadata.Company = &meta.Company
	out.Metadata.Email = &meta.Email
	out.Metadata.Name = &meta.Name
	out.Metadata.Description = &meta.Description
	out.Metadata.Location = &meta.Location
	out.Metadata.HasOrganizationProjects = &meta.HasOrganizationProjects
	out.Metadata.HasRepositoryProjects = &meta.HasRepositoryProjects
	drp := github.RepoPermissionLevel(meta.DefaultRepositoryPermission)
	out.Metadata.DefaultRepositoryPermission = &drp
	out.Metadata.MembersCanCreateRepositories = &meta.MembersCanCreateRepositories

	var runningAsAdmin bool
	runningAs, err := client.BotUser()
	if err != nil {
		return nil, fmt.Errorf("failed to obtain username for this token")
	}
	admins, err := client.ListOrgMembers(orgName, github.RoleAdmin)
	if err != nil {
		return nil, fmt.Errorf("failed to list org admins: %w", err)
	}
	logrus.Debugf("Found %d admins", len(admins))
	for _, m := range admins {
		logrus.WithField("login", m.Login).Debug("Recording admin.")
		out.Admins = append(out.Admins, m.Login)
		if runningAs.Login == m.Login || appID != "" {
			runningAsAdmin = true
		}
	}

	if !runningAsAdmin {
		return nil, fmt.Errorf("--dump must be run with admin:org scope token")
	}

	orgMembers, err := client.ListOrgMembers(orgName, github.RoleMember)
	if err != nil {
		return nil, fmt.Errorf("failed to list org members: %w", err)
	}
	logrus.Debugf("Found %d members", len(orgMembers))
	for _, m := range orgMembers {
		logrus.WithField("login", m.Login).Debug("Recording member.")
		out.Members = append(out.Members, m.Login)
	}

	teams, err := client.ListTeams(orgName)
	if err != nil {
		return nil, fmt.Errorf("failed to list teams: %w", err)
	}
	logrus.Debugf("Found %d teams", len(teams))

	names := map[int]string{}   // what's the name of a team?
	idMap := map[int]org.Team{} // metadata for a team
	children := map[int][]int{} // what children does it have
	var tops []int              // what are the top-level teams

	for _, t := range teams {
		logger := logrus.WithFields(logrus.Fields{"id": t.ID, "name": t.Name})
		p := org.Privacy(t.Privacy)
		if ignoreSecretTeams && p == org.Secret {
			logger.Debug("Ignoring secret team.")
			continue
		}
		d := t.Description
		nt := org.Team{
			TeamMetadata: org.TeamMetadata{
				Description: &d,
				Privacy:     &p,
			},
			Maintainers: []string{},
			Members:     []string{},
			Children:    map[string]org.Team{},
			Repos:       map[string]github.RepoPermissionLevel{},
		}
		maintainers, err := client.ListTeamMembersBySlug(orgName, t.Slug, github.RoleMaintainer)
		if err != nil {
			return nil, fmt.Errorf("failed to list team %d(%s) maintainers: %w", t.ID, t.Name, err)
		}
		logger.Debugf("Found %d maintainers.", len(maintainers))
		for _, m := range maintainers {
			logger.WithField("login", m.Login).Debug("Recording maintainer.")
			nt.Maintainers = append(nt.Maintainers, m.Login)
		}
		teamMembers, err := client.ListTeamMembersBySlug(orgName, t.Slug, github.RoleMember)
		if err != nil {
			return nil, fmt.Errorf("failed to list team %d(%s) members: %w", t.ID, t.Name, err)
		}
		logger.Debugf("Found %d members.", len(teamMembers))
		for _, m := range teamMembers {
			logger.WithField("login", m.Login).Debug("Recording member.")
			nt.Members = append(nt.Members, m.Login)
		}

		names[t.ID] = t.Name
		idMap[t.ID] = nt

		if t.Parent == nil { // top level team
			logger.Debug("Marking as top-level team.")
			tops = append(tops, t.ID)
		} else { // add this id to the list of the parent's children
			logger.Debugf("Marking as child team of %d.", t.Parent.ID)
			children[t.Parent.ID] = append(children[t.Parent.ID], t.ID)
		}

		repos, err := client.ListTeamReposBySlug(orgName, t.Slug)
		if err != nil {
			return nil, fmt.Errorf("failed to list team %d(%s) repos: %w", t.ID, t.Name, err)
		}
		logger.Debugf("Found %d repo permissions.", len(repos))
		for _, repo := range repos {
			level := github.LevelFromPermissions(repo.Permissions)
			logger.WithFields(logrus.Fields{"repo": repo, "permission": level}).Debug("Recording repo permission.")
			nt.Repos[repo.Name] = level
		}
	}

	var makeChild func(id int) org.Team
	makeChild = func(id int) org.Team {
		t := idMap[id]
		for _, cid := range children[id] {
			child := makeChild(cid)
			t.Children[names[cid]] = child
		}
		return t
	}

	out.Teams = make(map[string]org.Team, len(tops))
	for _, id := range tops {
		out.Teams[names[id]] = makeChild(id)
	}

	repos, err := client.GetRepos(orgName, false)
	if err != nil {
		return nil, fmt.Errorf("failed to list org repos: %w", err)
	}
	logrus.Debugf("Found %d repos", len(repos))
	out.Repos = make(map[string]org.Repo, len(repos))
	for _, repo := range repos {
		full, err := client.GetRepo(orgName, repo.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to get repo: %w", err)
		}
		logrus.WithField("repo", full.FullName).Debug("Recording repo.")
		out.Repos[full.Name] = org.PruneRepoDefaults(org.Repo{
			Description:      &full.Description,
			HomePage:         &full.Homepage,
			Private:          &full.Private,
			HasIssues:        &full.HasIssues,
			HasProjects:      &full.HasProjects,
			HasWiki:          &full.HasWiki,
			AllowMergeCommit: &full.AllowMergeCommit,
			AllowSquashMerge: &full.AllowSquashMerge,
			AllowRebaseMerge: &full.AllowRebaseMerge,
			Archived:         &full.Archived,
			DefaultBranch:    &full.DefaultBranch,
		})
	}

	return &out, nil
}
