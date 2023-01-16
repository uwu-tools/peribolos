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

	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/test-infra/prow/config/org"
	"k8s.io/test-infra/prow/github"

	"github.com/relengfam/peribolos/options/root"
)

type teamClient interface {
	ListTeams(org string) ([]github.Team, error)
	CreateTeam(org string, team github.Team) (*github.Team, error)
	DeleteTeamBySlug(org, teamSlug string) error
}

// configureTeams returns the ids for all expected team names, creating/deleting teams as necessary.
func configureTeams(client teamClient, orgName string, orgConfig org.Config, maxDelta float64, ignoreSecretTeams bool) (map[string]github.Team, error) {
	if err := validateTeamNames(orgConfig); err != nil {
		return nil, err
	}

	// What teams exist?
	teams := map[string]github.Team{}
	slugs := sets.String{}
	teamList, err := client.ListTeams(orgName)
	if err != nil {
		return nil, fmt.Errorf("failed to list teams: %w", err)
	}
	logrus.Debugf("Found %d teams", len(teamList))
	for _, t := range teamList {
		if ignoreSecretTeams && org.Privacy(t.Privacy) == org.Secret {
			continue
		}
		teams[t.Slug] = t
		slugs.Insert(t.Slug)
	}
	if ignoreSecretTeams {
		logrus.Debugf("Found %d non-secret teams", len(teamList))
	}

	// What is the lowest ID for each team?
	older := map[string][]github.Team{}
	names := map[string]github.Team{}
	for _, t := range teams {
		logger := logrus.WithFields(logrus.Fields{"id": t.ID, "name": t.Name})
		n := t.Name
		switch val, ok := names[n]; {
		case !ok: // first occurrence of the name
			logger.Debug("First occurrence of this team name.")
			names[n] = t
		case ok && t.ID < val.ID: // t has the lower ID, replace and send current to older set
			logger.Debugf("Replacing previous recorded team (%d) with this one due to smaller ID.", val.ID)
			names[n] = t
			older[n] = append(older[n], val)
		default: // t does not have smallest id, add it to older set
			logger.Debugf("Adding team (%d) to older set as a smaller ID is already recoded for it.", val.ID)
			older[n] = append(older[n], val)
		}
	}

	// What team are we using for each configured name, and which names are missing?
	matches := map[string]github.Team{}
	missing := map[string]org.Team{}
	used := sets.String{}
	var match func(teams map[string]org.Team)
	match = func(teams map[string]org.Team) {
		for name, orgTeam := range teams {
			logger := logrus.WithField("name", name)
			match(orgTeam.Children)
			t := findTeam(names, name, orgTeam.Previously...)
			if t == nil {
				missing[name] = orgTeam
				logger.Debug("Could not find team in GitHub for this configuration.")
				continue
			}
			matches[name] = *t // t.Name != name if we matched on orgTeam.Previously
			logger.WithField("id", t.ID).Debug("Found a team in GitHub for this configuration.")
			used.Insert(t.Slug)
		}
	}
	match(orgConfig.Teams)

	// First compute teams we will delete, ensure we are not deleting too many
	unused := slugs.Difference(used)
	if delta := float64(len(unused)) / float64(len(slugs)); delta > maxDelta {
		return nil, fmt.Errorf("cannot delete %d teams or %.3f of %s teams (exceeds limit of %.3f)", len(unused), delta, orgName, maxDelta)
	}

	// Create any missing team names
	var failures []string
	for name, orgTeam := range missing {
		t := &github.Team{Name: name}
		if orgTeam.Description != nil {
			t.Description = *orgTeam.Description
		}
		if orgTeam.Privacy != nil {
			t.Privacy = string(*orgTeam.Privacy)
		}
		t, err := client.CreateTeam(orgName, *t)
		if err != nil {
			logrus.WithError(err).Warnf("Failed to create %s in %s", name, orgName)
			failures = append(failures, name)
			continue
		}
		matches[name] = *t
		// t.Slug may include a slug already present in slugs if other actors are deleting teams.
		used.Insert(t.Slug)
	}
	if n := len(failures); n > 0 {
		return nil, fmt.Errorf("failed to create %d teams: %s", n, strings.Join(failures, ", "))
	}

	// Remove any IDs returned by CreateTeam() that are in the unused set.
	if reused := unused.Intersection(used); len(reused) > 0 {
		// Logically possible for:
		// * another actor to delete team N after the ListTeams() call
		// * github to reuse team N after someone deleted it
		// Therefore used may now include IDs in unused, handle this situation.
		logrus.Warnf("Will not delete %d team IDs reused by github: %v", len(reused), reused.List())
		unused = unused.Difference(reused)
	}
	// Delete undeclared teams.
	for slug := range unused {
		if err := client.DeleteTeamBySlug(orgName, slug); err != nil {
			str := fmt.Sprintf("%s(%s)", slug, teams[slug].Name)
			logrus.WithError(err).Warnf("Failed to delete team %s from %s", str, orgName)
			failures = append(failures, str)
		}
	}
	if n := len(failures); n > 0 {
		return nil, fmt.Errorf("failed to delete %d teams: %s", n, strings.Join(failures, ", "))
	}

	// Return matches
	return matches, nil
}

// validateTeamNames returns an error if any current/previous names are used multiple times in the config.
func validateTeamNames(orgConfig org.Config) error {
	// Does the config duplicate any team names?
	used := sets.String{}
	dups := sets.String{}
	for name, orgTeam := range orgConfig.Teams {
		if used.Has(name) {
			dups.Insert(name)
		} else {
			used.Insert(name)
		}
		for _, n := range orgTeam.Previously {
			if used.Has(n) {
				dups.Insert(n)
			} else {
				used.Insert(n)
			}
		}
	}
	if n := len(dups); n > 0 {
		return fmt.Errorf("%d duplicated names: %s", n, strings.Join(dups.List(), ", "))
	}
	return nil
}

// findTeam returns teams[n] for the first n in [name, previousNames, ...] that is in teams.
func findTeam(teams map[string]github.Team, name string, previousNames ...string) *github.Team {
	if t, ok := teams[name]; ok {
		return &t
	}
	for _, p := range previousNames {
		if t, ok := teams[p]; ok {
			return &t
		}
	}
	return nil
}

func configureTeamAndMembers(opt root.Options, client github.Client, githubTeams map[string]github.Team, name, orgName string, team org.Team, parent *int) error {
	gt, ok := githubTeams[name]
	if !ok { // configureTeams is buggy if this is the case
		return fmt.Errorf("%s not found in id list", name)
	}

	// Configure team metadata
	err := configureTeam(client, orgName, name, team, gt, parent)
	if err != nil {
		return fmt.Errorf("failed to update %s metadata: %w", name, err)
	}

	// Configure team members
	if !opt.FixTeamMembers {
		logrus.Infof("Skipping %s member configuration", name)
	} else if err = configureTeamMembers(client, orgName, gt, team, opt.IgnoreInvitees); err != nil {
		return fmt.Errorf("failed to update %s members: %w", name, err)
	}

	for childName, childTeam := range team.Children {
		err = configureTeamAndMembers(opt, client, githubTeams, childName, orgName, childTeam, &gt.ID)
		if err != nil {
			return fmt.Errorf("failed to update %s child teams: %w", name, err)
		}
	}

	return nil
}

type editTeamClient interface {
	EditTeam(org string, team github.Team) (*github.Team, error)
}

// configureTeam patches the team name/description/privacy when values differ
func configureTeam(client editTeamClient, orgName, teamName string, team org.Team, gt github.Team, parent *int) error {
	// Do we need to reconfigure any team settings?
	patch := false
	if gt.Name != teamName {
		patch = true
	}
	gt.Name = teamName
	if team.Description != nil && gt.Description != *team.Description {
		patch = true
		gt.Description = *team.Description
	} else {
		gt.Description = ""
	}
	// doesn't have parent in github, but has parent in config
	if gt.Parent == nil && parent != nil {
		patch = true
		gt.ParentTeamID = parent
	}
	if gt.Parent != nil { // has parent in github ...
		if parent == nil { // ... but doesn't need one
			patch = true
			gt.Parent = nil
			gt.ParentTeamID = parent
		} else if gt.Parent.ID != *parent { // but it's different than the config
			patch = true
			gt.Parent = nil
			gt.ParentTeamID = parent
		}
	}

	if team.Privacy != nil && gt.Privacy != string(*team.Privacy) {
		patch = true
		gt.Privacy = string(*team.Privacy)

	} else if team.Privacy == nil && (parent != nil || len(team.Children) > 0) && gt.Privacy != "closed" {
		patch = true
		gt.Privacy = github.PrivacyClosed // nested teams must be closed
	}

	if patch { // yes we need to patch
		if _, err := client.EditTeam(orgName, gt); err != nil {
			return fmt.Errorf("failed to edit %s team %s(%s): %w", orgName, gt.Slug, gt.Name, err)
		}
	}
	return nil
}

// teamMembersClient can list/remove/update people to a team.
type teamMembersClient interface {
	ListTeamMembersBySlug(org, teamSlug, role string) ([]github.TeamMember, error)
	ListTeamInvitationsBySlug(org, teamSlug string) ([]github.OrgInvitation, error)
	RemoveTeamMembershipBySlug(org, teamSlug, user string) error
	UpdateTeamMembershipBySlug(org, teamSlug, user string, maintainer bool) (*github.TeamMembership, error)
}

// configureTeamMembers will add/update people to the appropriate role on the team, and remove anyone else.
func configureTeamMembers(client teamMembersClient, orgName string, gt github.Team, team org.Team, ignoreInvitees bool) error {
	// Get desired state
	wantMaintainers := sets.NewString(team.Maintainers...)
	wantMembers := sets.NewString(team.Members...)

	// Get current state
	haveMaintainers := sets.String{}
	haveMembers := sets.String{}

	members, err := client.ListTeamMembersBySlug(orgName, gt.Slug, github.RoleMember)
	if err != nil {
		return fmt.Errorf("failed to list %s(%s) members: %w", gt.Slug, gt.Name, err)
	}
	for _, m := range members {
		haveMembers.Insert(m.Login)
	}

	maintainers, err := client.ListTeamMembersBySlug(orgName, gt.Slug, github.RoleMaintainer)
	if err != nil {
		return fmt.Errorf("failed to list %s(%s) maintainers: %w", gt.Slug, gt.Name, err)
	}
	for _, m := range maintainers {
		haveMaintainers.Insert(m.Login)
	}

	invitees := sets.String{}
	if !ignoreInvitees {
		invitees, err = teamInvitations(client, orgName, gt.Slug)
		if err != nil {
			return fmt.Errorf("failed to list %s(%s) invitees: %w", gt.Slug, gt.Name, err)
		}
	}

	adder := func(user string, super bool) error {
		if invitees.Has(user) {
			logrus.Infof("Waiting for %s to accept invitation to %s(%s)", user, gt.Slug, gt.Name)
			return nil
		}
		role := github.RoleMember
		if super {
			role = github.RoleMaintainer
		}
		tm, err := client.UpdateTeamMembershipBySlug(orgName, gt.Slug, user, super)
		if err != nil {
			// Augment the error with the operation we attempted so that the error makes sense after return
			err = fmt.Errorf("UpdateTeamMembership(%s(%s), %s, %t) failed: %w", gt.Slug, gt.Name, user, super, err)
			logrus.Warnf(err.Error())
		} else if tm.State == github.StatePending {
			logrus.Infof("Invited %s to %s(%s) as a %s", user, gt.Slug, gt.Name, role)
		} else {
			logrus.Infof("Set %s as a %s of %s(%s)", user, role, gt.Slug, gt.Name)
		}
		return err
	}

	remover := func(user string) error {
		err := client.RemoveTeamMembershipBySlug(orgName, gt.Slug, user)
		if err != nil {
			// Augment the error with the operation we attempted so that the error makes sense after return
			err = fmt.Errorf("RemoveTeamMembership(%s(%s), %s) failed: %w", gt.Slug, gt.Name, user, err)
			logrus.Warnf(err.Error())
		} else {
			logrus.Infof("Removed %s from team %s(%s)", user, gt.Slug, gt.Name)
		}
		return err
	}

	want := memberships{members: wantMembers, super: wantMaintainers}
	have := memberships{members: haveMembers, super: haveMaintainers}
	return configureMembers(have, want, invitees, adder, remover)
}

func teamInvitations(client teamMembersClient, orgName, teamSlug string) (sets.String, error) {
	invitees := sets.String{}
	is, err := client.ListTeamInvitationsBySlug(orgName, teamSlug)
	if err != nil {
		return nil, err
	}
	for _, i := range is {
		if i.Login == "" {
			continue
		}
		invitees.Insert(github.NormLogin(i.Login))
	}
	return invitees, nil
}

type teamRepoClient interface {
	ListTeamReposBySlug(org, teamSlug string) ([]github.Repo, error)
	UpdateTeamRepoBySlug(org, teamSlug, repo string, permission github.TeamPermission) error
	RemoveTeamRepoBySlug(org, teamSlug, repo string) error
}

// configureTeamRepos updates the list of repos that the team has permissions for when necessary
func configureTeamRepos(client teamRepoClient, githubTeams map[string]github.Team, name, orgName string, team org.Team) error {
	gt, ok := githubTeams[name]
	if !ok { // configureTeams is buggy if this is the case
		return fmt.Errorf("%s not found in id list", name)
	}

	want := team.Repos
	have := map[string]github.RepoPermissionLevel{}
	repos, err := client.ListTeamReposBySlug(orgName, gt.Slug)
	if err != nil {
		return fmt.Errorf("failed to list team %d(%s) repos: %w", gt.ID, name, err)
	}
	for _, repo := range repos {
		have[repo.Name] = github.LevelFromPermissions(repo.Permissions)
	}

	actions := map[string]github.RepoPermissionLevel{}
	for wantRepo, wantPermission := range want {
		if havePermission, haveRepo := have[wantRepo]; haveRepo && havePermission == wantPermission {
			// nothing to do
			continue
		}
		// create or update this permission
		actions[wantRepo] = wantPermission
	}

	for haveRepo := range have {
		if _, wantRepo := want[haveRepo]; !wantRepo {
			// should remove these permissions
			actions[haveRepo] = github.None
		}
	}

	var updateErrors []error
	for repo, permission := range actions {
		var err error
		switch permission {
		case github.None:
			err = client.RemoveTeamRepoBySlug(orgName, gt.Slug, repo)
		case github.Admin:
			err = client.UpdateTeamRepoBySlug(orgName, gt.Slug, repo, github.RepoAdmin)
		case github.Write:
			err = client.UpdateTeamRepoBySlug(orgName, gt.Slug, repo, github.RepoPush)
		case github.Read:
			err = client.UpdateTeamRepoBySlug(orgName, gt.Slug, repo, github.RepoPull)
		case github.Triage:
			err = client.UpdateTeamRepoBySlug(orgName, gt.Slug, repo, github.RepoTriage)
		case github.Maintain:
			err = client.UpdateTeamRepoBySlug(orgName, gt.Slug, repo, github.RepoMaintain)
		}

		if err != nil {
			updateErrors = append(updateErrors, fmt.Errorf("failed to update team %d(%s) permissions on repo %s to %s: %w", gt.ID, name, repo, permission, err))
		}
	}

	for childName, childTeam := range team.Children {
		if err := configureTeamRepos(client, githubTeams, childName, orgName, childTeam); err != nil {
			updateErrors = append(updateErrors, fmt.Errorf("failed to configure %s child team %s repos: %w", orgName, childName, err))
		}
	}

	return utilerrors.NewAggregate(updateErrors)
}
