# Overview

A really simple CI/CD server.  It listens for webhooks from Github, and when it receives one for a configured repository it runs the commands that you've specified.   No abstractions, no elaborate interfaces.  A single binary, a single basic config file and you're off.


# Approach

You MUST make a plan before starting work and you MUST get that plan approved by asking the human.

Each item of work will take place on a new branch and the human will ensure the branch is created BEFORE he gives you a prompt. He will do this using his "gitgo" command which ensures there is ALWAYS a fresh copy of main that is up to date BEFORE a new branch is created.

You MUST ensure that that tests pass for each step of your plan before moving on to the next part. You MUST ALWAYS fix broken tests by fixing non-test code but if you think the tests need to be edited, you MUST get permission from the human.

When one part of your plan is complete, you MUST commit your work using a conventional commit message.

When you have completed your work, create a PR on github and summarise the work.

The title of the pr MUST be in the format of a convention commit message. As this is used for squash merging the code.

NEVER unstage unstage anything in .prompts as I use this directory to save the prompts you are given.

NEVER unstage version.txt. I edit that file as we work together.
