// Copyright 2023 uwu-tools Authors
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

// Derived from https://github.com/airconduct/go-probot/blob/main/example/main.go.
package main

import (
	"context"

	"github.com/airconduct/go-probot"
	probotgh "github.com/airconduct/go-probot/github"
	"github.com/google/go-github/v48/github"
	"github.com/spf13/pflag"
)

func main() {
	app := probot.NewGithubAPP()
	app.AddFlags(pflag.CommandLine)
	pflag.Parse()

	// Add a handler for events "issue_comment.created"
	app.On(probotgh.Event.IssueComment_created).WithHandler(
		probotgh.IssueCommentHandler(
			func(ctx probotgh.IssueCommentContext) {
				payload := ctx.Payload()
				ctx.Logger().Info("Get IssueComment event", "payload", payload)
				owner := payload.Repo.Owner.GetLogin()
				repo := payload.Repo.GetName()
				issueNumber := *payload.Issue.Number

				// If any error happen, the error message will be logged and sent as response
				ctx.Must(
					ctx.Client().Issues.CreateComment(
						ctx, owner, repo, issueNumber, &github.IssueComment{
							Body: github.String("Reply to this comment."),
						},
					),
				)
			},
		),
	)

	// Add a handler for multiple events
	app.On(
		probotgh.Event.PullRequest_opened,      // pull_request.opened
		probotgh.Event.PullRequest_edited,      // pull_request.edited
		probotgh.Event.PullRequest_synchronize, // pull_request.synchronize
		probotgh.Event.PullRequest_labeled,     // pull_request.labeled
		probotgh.Event.PullRequest_assigned,    // pull_request.assigned
	).WithHandler(probotgh.PullRequestHandler(
		func(ctx probotgh.PullRequestContext) {
			payload := ctx.Payload()
			ctx.Logger().Info(
				"Do something",
				"action",
				payload.GetAction(),
				"PullRequest labels",
				payload.PullRequest.Labels,
			)
		},
	))

	if err := app.Run(context.Background()); err != nil {
		panic(err)
	}
}
