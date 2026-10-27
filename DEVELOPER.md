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

## Running SDKv2 hello reference provider acceptance test

```sh
TF_ACC=1 go -C internal/reference-provider test ./... -v -count=1
```

To run selected tests based on test name regex pattern:

```sh
TF_ACC=1 go -C internal/reference-provider test ./... -v -count=1 -run="^TestAccHotAirBalloonInvalidAdditionalGasTankAllowedIfValueIsUnknown"
```

## Other useful commands

See `uv run --locked --project internal/cli svp --help`