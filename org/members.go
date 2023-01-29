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

	"github.com/uwu-tools/peribolos/options/root"
)

type inviteClient interface {
	ListOrgInvitations(org string) ([]github.OrgInvitation, error)
}

func orgInvitations(opt root.Options, client inviteClient, orgName string) (sets.String, error) {
	invitees := sets.String{}
	if !opt.FixOrgMembers && !opt.FixTeamMembers {
		return invitees, nil
	}
	is, err := client.ListOrgInvitations(orgName)
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

type orgClient interface {
	BotUser() (*github.UserData, error)
	ListOrgMembers(org, role string) ([]github.TeamMember, error)
	RemoveOrgMembership(org, user string) error
	UpdateOrgMembership(org, user string, admin bool) (*github.OrgMembership, error)
}

func configureOrgMembers(opt root.Options, client orgClient, orgName string, orgConfig org.Config, invitees sets.String) error {
	// Get desired state
	wantAdmins := sets.NewString(orgConfig.Admins...)
	wantMembers := sets.NewString(orgConfig.Members...)

	// Sanity desired state
	if n := len(wantAdmins); n < opt.MinAdmins {
		return fmt.Errorf("%s must specify at least %d admins, only found %d", orgName, opt.MinAdmins, n)
	}
	var missing []string
	for _, r := range opt.RequiredAdmins {
		if !wantAdmins.Has(r) {
			missing = append(missing, r)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("%s must specify %v as admins, missing %v", orgName, opt.RequiredAdmins, missing)
	}
	if opt.RequireSelf {
		if me, err := client.BotUser(); err != nil {
			return fmt.Errorf("cannot determine user making requests for %s: %v", opt.GithubOpts.TokenPath, err)
		} else if !wantAdmins.Has(me.Login) {
			return fmt.Errorf("authenticated user %s is not an admin of %s", me.Login, orgName)
		}
	}

	// Get current state
	haveAdmins := sets.String{}
	haveMembers := sets.String{}
	ms, err := client.ListOrgMembers(orgName, github.RoleAdmin)
	if err != nil {
		return fmt.Errorf("failed to list %s admins: %w", orgName, err)
	}
	for _, m := range ms {
		haveAdmins.Insert(m.Login)
	}
	if ms, err = client.ListOrgMembers(orgName, github.RoleMember); err != nil {
		return fmt.Errorf("failed to list %s members: %w", orgName, err)
	}
	for _, m := range ms {
		haveMembers.Insert(m.Login)
	}

	have := memberships{members: haveMembers, super: haveAdmins}
	want := memberships{members: wantMembers, super: wantAdmins}
	have.normalize()
	want.normalize()
	// Figure out who to remove
	remove := have.all().Difference(want.all())

	// Sanity check changes
	if d := float64(len(remove)) / float64(len(have.all())); d > opt.MaxDelta {
		return fmt.Errorf("cannot delete %d memberships or %.3f of %s (exceeds limit of %.3f)", len(remove), d, orgName, opt.MaxDelta)
	}

	teamMembers := sets.String{}
	teamNames := sets.String{}
	duplicateTeamNames := sets.String{}
	for name, team := range orgConfig.Teams {
		teamMembers.Insert(team.Members...)
		teamMembers.Insert(team.Maintainers...)
		if teamNames.Has(name) {
			duplicateTeamNames.Insert(name)
		}
		teamNames.Insert(name)
		for _, n := range team.Previously {
			if teamNames.Has(n) {
				duplicateTeamNames.Insert(n)
			}
			teamNames.Insert(n)
		}
	}

	teamMembers = normalize(teamMembers)
	if outside := teamMembers.Difference(want.all()); len(outside) > 0 {
		return fmt.Errorf("all team members/maintainers must also be org members: %s", strings.Join(outside.List(), ", "))
	}

	if n := len(duplicateTeamNames); n > 0 {
		return fmt.Errorf("team names must be unique (including previous names), %d duplicated names: %s", n, strings.Join(duplicateTeamNames.List(), ", "))
	}

	adder := func(user string, super bool) error {
		if invitees.Has(user) { // Do not add them, as this causes another invite.
			logrus.Infof("Waiting for %s to accept invitation to %s", user, orgName)
			return nil
		}
		role := github.RoleMember
		if super {
			role = github.RoleAdmin
		}
		om, err := client.UpdateOrgMembership(orgName, user, super)
		if err != nil {
			logrus.WithError(err).Warnf("UpdateOrgMembership(%s, %s, %t) failed", orgName, user, super)
			if github.IsNotFound(err) {
				// this could be caused by someone removing their account
				// or a typo in the configuration but should not crash the sync
				err = nil
			}
		} else if om.State == github.StatePending {
			logrus.Infof("Invited %s to %s as a %s", user, orgName, role)
		} else {
			logrus.Infof("Set %s as a %s of %s", user, role, orgName)
		}
		return err
	}

	remover := func(user string) error {
		err := client.RemoveOrgMembership(orgName, user)
		if err != nil {
			logrus.WithError(err).Warnf("RemoveOrgMembership(%s, %s) failed", orgName, user)
		}
		return err
	}

	return configureMembers(have, want, invitees, adder, remover)
}

type memberships struct {
	members sets.String
	super   sets.String
}

func (m memberships) all() sets.String {
	return m.members.Union(m.super)
}

func (m *memberships) normalize() {
	m.members = normalize(m.members)
	m.super = normalize(m.super)
}

func configureMembers(have, want memberships, invitees sets.String, adder func(user string, super bool) error, remover func(user string) error) error {
	have.normalize()
	want.normalize()
	if both := want.super.Intersection(want.members); len(both) > 0 {
		return fmt.Errorf("users in both roles: %s", strings.Join(both.List(), ", "))
	}
	havePlusInvites := have.all().Union(invitees)
	remove := havePlusInvites.Difference(want.all())
	members := want.members.Difference(have.members)
	supers := want.super.Difference(have.super)

	var errs []error
	for u := range members {
		if err := adder(u, false); err != nil {
			errs = append(errs, err)
		}
	}
	for u := range supers {
		if err := adder(u, true); err != nil {
			errs = append(errs, err)
		}
	}

	for u := range remove {
		if err := remover(u); err != nil {
			errs = append(errs, err)
		}
	}

	return utilerrors.NewAggregate(errs)
}
