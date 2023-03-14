// Copyright 2022 RelEngFam Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package root

import (
	"flag"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/uwu-tools/peribolos/third_party/k8s.io/test-infra/prow/flagutil"
)

const (
	// Flags.

	// Configuration settings.
	flagConfigPath = "config-path"
	flagConfirm    = "confirm"
	flagDump       = "dump"
	flagDumpFull   = "dump-full"
	flagLogLevel   = "log-level"

	// Protections.
	// TODO(action): Missing input parameter
	flagMaxRemovalDelta = "maximum-removal-delta"
	flagMinAdmins       = "min-admins"
	flagRequireSelf     = "require-self"
	// TODO(action): Missing input parameter
	flagRequiredAdmins = "required-admins"

	// Organization settings.
	flagFixOrg = "fix-org"
	// TODO(action): Missing input parameter
	flagIgnoreInvitees = "ignore-invitees"

	// Members settings.
	flagFixOrgMembers = "fix-org-members"

	// Team settings.
	flagFixTeams       = "fix-teams"
	flagFixTeamMembers = "fix-team-members"
	flagFixTeamRepos   = "fix-team-repos"
	// TODO(action): Missing input parameter
	flagIgnoreSecretTeams = "ignore-secret-teams"

	// Repo settings.
	flagFixRepos = "fix-repos"
	// TODO(action): Missing input parameter
	flagAllowRepoArchival = "allow-repo-archival"
	// TODO(action): Missing input parameter
	flagAllowRepoPublish = "allow-repo-publish"

	// Prow GitHub settings.
	// TODO(action): Missing input parameter
	flagTokens = "tokens"
	// TODO(action): Missing input parameter
	flagTokenBurst = "token-burst"
)

// Command is an interface for handling options for command-line utilities.
type Command interface {
	// AddFlags adds this options' flags to the cobra command.
	AddFlags(cmd *cobra.Command)
}

// AddFlags adds this options' flags to the cobra command.
func (o *Options) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringSliceVar(
		&o.RequiredAdmins,
		flagRequiredAdmins,
		o.RequiredAdmins,
		"Ensure config specifies these users as admins",
	)

	cmd.Flags().IntVar(
		&o.MinAdmins,
		flagMinAdmins,
		defaultMinAdmins,
		"Ensure config specifies at least this many admins",
	)

	cmd.Flags().BoolVar(
		&o.RequireSelf,
		flagRequireSelf,
		true,
		"Ensure github token path user is an admin",
	)

	cmd.Flags().Float64Var(
		&o.MaxDelta,
		flagMaxRemovalDelta,
		defaultDelta,
		"Fail if config removes more than this fraction of current members",
	)

	cmd.Flags().StringVar(
		&o.Config,
		flagConfigPath,
		"",
		"Path to org config.yaml",
	)

	cmd.Flags().BoolVar(
		&o.Confirm,
		flagConfirm,
		false,
		"Mutate github if set",
	)

	cmd.Flags().StringVar(
		&o.Dump,
		flagDump,
		"",
		"Output current config of this org if set",
	)

	cmd.Flags().BoolVar(
		&o.DumpFull,
		flagDumpFull,
		false,
		"Output current config of the org as a valid input config file instead of a snippet",
	)

	cmd.Flags().BoolVar(
		&o.IgnoreInvitees,
		flagIgnoreInvitees,
		false,
		"Do not compare missing members with active invitations (compatibility for GitHub Enterprise)",
	)

	cmd.Flags().BoolVar(
		&o.IgnoreSecretTeams,
		flagIgnoreSecretTeams,
		false,
		"Do not dump or update secret teams if set",
	)

	cmd.Flags().BoolVar(
		&o.FixOrg,
		flagFixOrg,
		false,
		"Change org metadata if set",
	)

	cmd.Flags().BoolVar(
		&o.FixOrgMembers,
		flagFixOrgMembers,
		false,
		"Add/remove org members if set",
	)

	cmd.Flags().BoolVar(
		&o.FixTeams,
		flagFixTeams,
		false,
		"Create/delete/update teams if set",
	)

	cmd.Flags().BoolVar(
		&o.FixTeamMembers,
		flagFixTeamMembers,
		false,
		"Add/remove team members if set",
	)

	cmd.Flags().BoolVar(
		&o.FixTeamRepos,
		flagFixTeamRepos,
		false,
		"Add/remove team permissions on repos if set",
	)

	cmd.Flags().BoolVar(
		&o.FixRepos,
		flagFixRepos,
		false,
		"Create/update repositories if set",
	)

	cmd.Flags().BoolVar(
		&o.AllowRepoArchival,
		flagAllowRepoArchival,
		false,
		"If set, archiving repos is allowed while updating repos",
	)

	cmd.Flags().BoolVar(
		&o.AllowRepoPublish,
		flagAllowRepoPublish,
		false,
		"If set, making private repos public is allowed while updating repos",
	)

	cmd.Flags().StringVar(
		&o.logLevel,
		flagLogLevel,
		logrus.InfoLevel.String(),
		fmt.Sprintf("Logging level, one of %v", logrus.AllLevels),
	)

	ghFlags := flag.NewFlagSet("github-flags", flag.ContinueOnError)
	o.GithubOpts.AddCustomizedFlags(ghFlags, flagutil.ThrottlerDefaults(defaultTokens, defaultBurst))

	cmd.Flags().AddGoFlagSet(ghFlags)
}
