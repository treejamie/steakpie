# Overview

Steak Pie is a Go service that listens for GitHub registry package webhooks. When a new container image is pushed to ghcr.io, it automatically pulls the latest version and updates your local Docker Compose stack.


# Approach

You MUST make a plan before starting work and you MUST get that plan approved by asking Jamie the human.

Each item of work will take place on a new branch and Jamie the human will ensure the branch is created BEFORE he gives you a prompt. He will do this using his "gitgo" command which ensures there is ALWAYS a fresh copy of main that is up to date BEFORE a new branch is created.

You MUST ensure that that tests pass for each step of your plan before moving on to the next part. You MUST ALWAYS fix broken tests by fixing non-test code but if you think the tests need to be edited, you MUST get permission from Jamie the human.

When one part of your plan is complete, you MUST commit your work using a conventional commit message.

When you have completed your work, create a PR on github and summarise the work.

The title of the pr MUST be in the format of a convention commit message. As this is used for squash merging the code.

I'm adding prompt files as I go along, so dont' unstage anything in .prompts.