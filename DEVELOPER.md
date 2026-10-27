# Developer Guide

## Pre-requisites

1. [uv](https://docs.astral.sh/uv/getting-started/installation/)
2. [go](https://golang.org/doc/install)
3. [tfenv](https://github.com/tfutils/tfenv#installation)

## Pre-commit hook

Install the pre-commit hook for early validation of code changes.

```sh
uv run --locked --project internal/cli svp install-git-precommit
```