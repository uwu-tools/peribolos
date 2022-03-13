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

package options

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/caarlos0/env/v6"
	actions "github.com/sethvargo/go-githubactions"
	"github.com/sirupsen/logrus"

	"k8s.io/test-infra/prow/flagutil"
)

const (
	defaultMinAdmins = 5
	defaultDelta     = 0.25
	defaultTokens    = 300
	defaultBurst     = 100

	// Flags.
	flagRequiredAdmins    = "required-admins"
	flagMinAdmins         = "min-admins"
	flagRequireSelf       = "require-self"
	flagMaxRemovalDelta   = "maximum-removal-delta"
	flagConfigPath        = "config-path"
	flagConfirm           = "confirm"
	flagTokens            = "tokens"
	flagTokenBurst        = "token-burst"
	flagDump              = "dump"
	flagDumpFull          = "dump-full"
	flagIgnoreSecretTeams = "ignore-secret-teams"
	flagFixOrg            = "fix-org"
	flagFixOrgMembers     = "fix-org-members"
	flagFixTeams          = "fix-teams"
	flagFixTeamMembers    = "fix-team-members"
	flagFixTeamRepos      = "fix-team-repos"
	flagFixRepos          = "fix-repos"
	flagAllowRepoArchival = "allow-repo-archival"
	flagAllowRepoPublish  = "allow-repo-publish"
	flagLogLevel          = "log-level"
)

type Options struct {
	// Infer if peribolos is running in a GitHub Action.
	// The `CI` environment variable will always be set to "true" in a GitHub Action.
	// ref: https://docs.github.com/en/actions/learn-github-actions/environment-variables#default-environment-variables
	UsingActions      bool `env:"CI"`
	Config            string
	Confirm           bool
	Dump              string
	DumpFull          bool
	MaxDelta          float64
	MinAdmins         int
	RequireSelf       bool
	RequiredAdmins    flagutil.Strings
	FixOrg            bool
	FixOrgMembers     bool
	FixTeamMembers    bool
	FixTeams          bool
	FixTeamRepos      bool
	FixRepos          bool
	IgnoreSecretTeams bool
	AllowRepoArchival bool
	AllowRepoPublish  bool
	GithubOpts        flagutil.GitHubOptions

	// TODO(petr-muller): Remove after August 2021, replaced by github.ThrottleHourlyTokens
	tokenBurst    int
	tokensPerHour int

	logLevel string
}

func New() Options {
	o := new()
	if err := o.parseArgs(flag.CommandLine, os.Args[1:]); err != nil {
		logrus.Fatalf("Invalid flags: %v", err)
	}

	return *o
}

func new() *Options {
	o := &Options{}
	if err := env.Parse(o); err != nil {
		fmt.Printf("could not parse env vars, using default options: %v", err)
	}

	return o
}

func (o *Options) parseArgs(flags *flag.FlagSet, args []string) error {
	o.RequiredAdmins = flagutil.NewStrings()

	flags.Var(
		&o.RequiredAdmins,
		flagRequiredAdmins,
		"Ensure config specifies these users as admins",
	)

	flags.IntVar(
		&o.MinAdmins,
		flagMinAdmins,
		defaultMinAdmins,
		"Ensure config specifies at least this many admins",
	)

	flags.BoolVar(
		&o.RequireSelf,
		flagRequireSelf,
		true,
		"Ensure github token path user is an admin",
	)

	flags.Float64Var(
		&o.MaxDelta,
		flagMaxRemovalDelta,
		defaultDelta,
		"Fail if config removes more than this fraction of current members",
	)

	flags.StringVar(
		&o.Config,
		flagConfigPath,
		"",
		"Path to org config.yaml",
	)

	flags.BoolVar(
		&o.Confirm,
		flagConfirm,
		false,
		"Mutate github if set",
	)

	flags.IntVar(
		&o.tokensPerHour,
		flagTokens,
		defaultTokens,
		"Throttle hourly token consumption (0 to disable) DEPRECATED: use --github-hourly-tokens",
	)

	flags.IntVar(
		&o.tokenBurst,
		flagTokenBurst,
		defaultBurst,
		"Allow consuming a subset of hourly tokens in a short burst. DEPRECATED: use --github-allowed-burst",
	)

	flags.StringVar(
		&o.Dump,
		flagDump,
		"",
		"Output current config of this org if set",
	)

	flags.BoolVar(
		&o.DumpFull,
		flagDumpFull,
		false,
		"Output current config of the org as a valid input config file instead of a snippet",
	)

	flags.BoolVar(
		&o.IgnoreSecretTeams,
		flagIgnoreSecretTeams,
		false,
		"Do not dump or update secret teams if set",
	)

	flags.BoolVar(
		&o.FixOrg,
		flagFixOrg,
		false,
		"Change org metadata if set",
	)

	flags.BoolVar(
		&o.FixOrgMembers,
		flagFixOrgMembers,
		false,
		"Add/remove org members if set",
	)

	flags.BoolVar(
		&o.FixTeams,
		flagFixTeams,
		false,
		"Create/delete/update teams if set",
	)

	flags.BoolVar(
		&o.FixTeamMembers,
		flagFixTeamMembers,
		false,
		"Add/remove team members if set",
	)

	flags.BoolVar(
		&o.FixTeamRepos,
		flagFixTeamRepos,
		false,
		"Add/remove team permissions on repos if set",
	)

	flags.BoolVar(
		&o.FixRepos,
		flagFixRepos,
		false,
		"Create/update repositories if set",
	)

	flags.BoolVar(
		&o.AllowRepoArchival,
		flagAllowRepoArchival,
		false,
		"If set, archiving repos is allowed while updating repos",
	)

	flags.BoolVar(
		&o.AllowRepoPublish,
		flagAllowRepoPublish,
		false,
		"If set, making private repos public is allowed while updating repos",
	)

	flags.StringVar(
		&o.logLevel,
		flagLogLevel,
		logrus.InfoLevel.String(),
		fmt.Sprintf("Logging level, one of %v", logrus.AllLevels),
	)

	o.GithubOpts.AddCustomizedFlags(flags, flagutil.ThrottlerDefaults(defaultTokens, defaultBurst))
	if err := flags.Parse(args); err != nil {
		return err
	}

	// TODO(flags): Consider parameterizing flag.
	o.GithubOpts.ThrottleHourlyTokens = o.tokensPerHour

	// TODO(flags): Consider parameterizing flag.
	o.GithubOpts.ThrottleAllowBurst = o.tokenBurst

	// TODO(actions): Add test case
	if o.UsingActions {
		fmt.Printf("Running in GitHub Actions environment")
		err := o.parseFromAction()
		if err != nil {
			return fmt.Errorf("parsing from Action: %w", err)
		}
	}

	if err := o.GithubOpts.Validate(!o.Confirm); err != nil {
		return fmt.Errorf(err.Error())
	}

	if o.MinAdmins < 2 {
		return fmt.Errorf("--min-admins=%d must be at least 2", o.MinAdmins)
	}

	if o.MaxDelta > 1 || o.MaxDelta < 0 {
		return fmt.Errorf("--maximum-removal-delta=%f must be a non-negative number less than 1.0", o.MaxDelta)
	}

	if o.Confirm && o.Dump != "" {
		return fmt.Errorf("--confirm cannot be used with --dump=%s", o.Dump)
	}
	if o.Config == "" && o.Dump == "" {
		return fmt.Errorf("--config-path or --dump required")
	}
	if o.Config != "" && o.Dump != "" {
		return fmt.Errorf("--config-path=%s and --dump=%s cannot both be set", o.Config, o.Dump)
	}

	if o.DumpFull && o.Dump == "" {
		return fmt.Errorf("--dump-full can't be used without --dump")
	}

	if o.FixTeamMembers && !o.FixTeams {
		return fmt.Errorf("--fix-team-members requires --fix-teams")
	}

	if o.FixTeamRepos && !o.FixTeams {
		return fmt.Errorf("--fix-team-repos requires --fix-teams")
	}

	level, err := logrus.ParseLevel(o.logLevel)
	if err != nil {
		return fmt.Errorf("--log-level invalid: %s", err.Error())
	}
	logrus.SetLevel(level)

	return nil
}

func (o *Options) parseFromAction() error {
	o.Dump = actions.GetInput(flagDump)
	o.Config = actions.GetInput(flagConfigPath)

	dumpFull := actions.GetInput(flagDumpFull)
	if dumpFull != "" {
		o.DumpFull, _ = strconv.ParseBool(dumpFull)
	}

	confirm := actions.GetInput(flagConfirm)
	if confirm != "" {
		o.Confirm, _ = strconv.ParseBool(confirm)
	}

	// TODO(flags): Consider parameterizing flag.
	o.GithubOpts.TokenPath = actions.GetInput("github-token-path")
	if o.GithubOpts.TokenPath == "" {
		return fmt.Errorf("missing 'github-token-path'")
	}

	FixOrg := actions.GetInput(flagFixOrg)
	if FixOrg != "" {
		o.FixOrg, _ = strconv.ParseBool(FixOrg)
	}

	FixOrgMembers := actions.GetInput(flagFixOrgMembers)
	if FixOrgMembers != "" {
		o.FixOrgMembers, _ = strconv.ParseBool(FixOrgMembers)
	}

	FixTeams := actions.GetInput(flagFixTeams)
	if FixTeams != "" {
		o.FixTeams, _ = strconv.ParseBool(FixTeams)
	}

	FixTeamMembers := actions.GetInput(flagFixTeamMembers)
	if FixTeamMembers != "" {
		o.FixTeamMembers, _ = strconv.ParseBool(FixTeamMembers)
	}

	FixTeamRepos := actions.GetInput(flagFixTeamRepos)
	if FixTeamRepos != "" {
		o.FixTeamRepos, _ = strconv.ParseBool(FixTeamRepos)
	}

	FixRepos := actions.GetInput(flagFixRepos)
	if FixRepos != "" {
		o.FixRepos, _ = strconv.ParseBool(FixRepos)
	}

	o.MinAdmins = defaultMinAdmins
	MinAdmins := actions.GetInput(flagMinAdmins)
	if MinAdmins != "" {
		o.MinAdmins, _ = strconv.Atoi(MinAdmins)
	}

	RequireSelf := actions.GetInput(flagRequireSelf)
	if RequireSelf != "" {
		o.RequireSelf, _ = strconv.ParseBool(RequireSelf)
	}

	throttleHourlyTokens := actions.GetInput("github-hourly-tokens")
	if throttleHourlyTokens != "" {
		o.GithubOpts.ThrottleHourlyTokens, _ = strconv.Atoi(throttleHourlyTokens)
	}

	throttleAllowBurst := actions.GetInput("github-allowed-burst")
	if throttleHourlyTokens != "" {
		o.GithubOpts.ThrottleAllowBurst, _ = strconv.Atoi(throttleAllowBurst)
	}

	logLevel := actions.GetInput(flagLogLevel)
	if logLevel != "" {
		o.logLevel = logLevel
	}

	return nil
}
