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

// TODO(tests): Uncomment tests once Codecov is working.
/*
func TestOptions(t *testing.T) {
	cases := []struct {
		name     string
		args     []string
		expected *Options
	}{
		{
			name: "missing --config",
			args: []string{},
		},
		{
			name: "bad --github-endpoint",
			args: []string{"--config-path=foo", "--github-endpoint=ht!tp://:dumb"},
		},
		{
			name: "--MinAdmins too low",
			args: []string{"--config-path=foo", "--min-admins=1"},
		},
		{
			name: "--maximum-removal-delta too high",
			args: []string{"--config-path=foo", "--maximum-removal-delta=1.1"},
		},
		{
			name: "--maximum-removal-delta too low",
			args: []string{"--config-path=foo", "--maximum-removal-delta=-0.1"},
		},
		{
			name: "reject --dump-full without --dump",
			args: []string{"--config-path=foo", "--dump-full"},
		},
		{
			name: "maximal delta",
			args: []string{"--config-path=foo", "--maximum-removal-delta=1"},
			expected: &Options{
				Config:        "foo",
				MinAdmins:     defaultMinAdmins,
				RequireSelf:   true,
				MaxDelta:      1,
				tokensPerHour: defaultTokens,
				tokenBurst:    defaultBurst,
				logLevel:      "info",
			},
		},
		{
			name: "minimal delta",
			args: []string{"--config-path=foo", "--maximum-removal-delta=0"},
			expected: &Options{
				Config:        "foo",
				MinAdmins:     defaultMinAdmins,
				RequireSelf:   true,
				MaxDelta:      0,
				tokensPerHour: defaultTokens,
				tokenBurst:    defaultBurst,
				logLevel:      "info",
			},
		},
		{
			name: "minimal admins",
			args: []string{"--config-path=foo", "--min-admins=2"},
			expected: &Options{
				Config:        "foo",
				MinAdmins:     2,
				RequireSelf:   true,
				MaxDelta:      defaultDelta,
				tokensPerHour: defaultTokens,
				tokenBurst:    defaultBurst,
				logLevel:      "info",
			},
		},
		{
			name: "reject burst > tokens",
			args: []string{"--config-path=foo", "--tokens=10", "--token-burst=11"},
		},
		{
			name: "reject dump and confirm",
			args: []string{"--confirm", "--dump=frogger"},
		},
		{
			name: "reject dump and config-path",
			args: []string{"--config-path=foo", "--dump=frogger"},
		},
		{
			name: "reject --fix-team-members without --fix-teams",
			args: []string{"--config-path=foo", "--fix-team-members"},
		},
		{
			name: "allow legacy disabled throttle",
			args: []string{"--config-path=foo", "--tokens=0"},
			expected: &Options{
				Config:      "foo",
				MinAdmins:   defaultMinAdmins,
				RequireSelf: true,
				MaxDelta:    defaultDelta,
				tokenBurst:  defaultBurst,
				logLevel:    "info",
			},
		},
		{
			name: "allow dump without config",
			args: []string{"--dump=frogger"},
			expected: &Options{
				MinAdmins:     defaultMinAdmins,
				RequireSelf:   true,
				MaxDelta:      defaultDelta,
				tokensPerHour: defaultTokens,
				tokenBurst:    defaultBurst,
				Dump:          "frogger",
				logLevel:      "info",
			},
		},
		{
			name: "minimal",
			args: []string{"--config-path=foo"},
			expected: &Options{
				Config:        "foo",
				MinAdmins:     defaultMinAdmins,
				RequireSelf:   true,
				MaxDelta:      defaultDelta,
				tokensPerHour: defaultTokens,
				tokenBurst:    defaultBurst,
				logLevel:      "info",
			},
		},
		{
			name: "full",
			args: []string{"--config-path=foo", "--github-token-path=bar", "--github-endpoint=weird://url", "--confirm=true", "--require-self=false", "--tokens=5", "--token-burst=2", "--dump=", "--fix-org", "--fix-org-members", "--fix-teams", "--fix-team-members", "--log-level=debug"},
			expected: &Options{
				Config:         "foo",
				Confirm:        true,
				RequireSelf:    false,
				MinAdmins:      defaultMinAdmins,
				MaxDelta:       defaultDelta,
				tokensPerHour:  5,
				tokenBurst:     2,
				FixOrg:         true,
				FixOrgMembers:  true,
				FixTeams:       true,
				FixTeamMembers: true,
				logLevel:       "debug",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			flags := flag.NewFlagSet(tc.name, flag.ContinueOnError)
			var actual Options
			err := actual.parseArgs(flags, tc.args)
			actual.GithubOpts = flagutil.GitHubOptions{}
			switch {
			case err == nil && tc.expected == nil:
				t.Errorf("%s: failed to return an error", tc.name)
			case err != nil && tc.expected != nil:
				t.Errorf("%s: unexpected error: %v", tc.name, err)
			case tc.expected != nil && !reflect.DeepEqual(*tc.expected, actual):
				t.Errorf("%s: got incorrect options: %v", tc.name, cmp.Diff(actual, *tc.expected, cmp.AllowUnexported(Options{}, flagutil.Strings{}, flagutil.GitHubOptions{})))
			}
		})
	}
}
*/
