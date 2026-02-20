# spk

Workspace CLI for multi-repo development — keeps repositories in sync, manages AWS credentials at the workspace level, and provides dependency-aware builds.

## Install

```bash
brew tap Spark-Rewards/spk
brew install spk
```

## Quick Start

```bash
# Create a workspace
spk create workspace ~/SparkRewards

# Optionally set AWS config
spk create workspace ~/SparkRewards --aws-profile dev --aws-region us-east-1

# Add repos
cd ~/SparkRewards
spk use my-org/backend-api --build "npm install && npm run build"
spk use my-org/frontend --build "npm install && npm run build" --deps backend-api

# Daily workflow
spk login           # AWS SSO login
spk sync            # git pull all repos
spk build --all     # build in dependency order

# Manage
spk list            # show repos + branch + status
spk workspace       # show workspace info
spk env set KEY=VAL # workspace-wide env vars
```

## Commands

| Command | Description |
|---|---|
| `spk create workspace <path>` | Initialize a new workspace |
| `spk use <org/repo>` | Clone and register a repo |
| `spk sync [repo]` | Pull latest for all or one repo |
| `spk build [repo] --all` | Run build commands (dependency-aware) |
| `spk login` | AWS SSO login using workspace profile |
| `spk list` | List repos with branch and status |
| `spk workspace` | Show workspace details |
| `spk config set --org <name>` | Set global defaults |
| `spk env set/list/export` | Manage workspace env vars |
| `spk remove <repo>` | Unregister a repo from manifest |
| `spk version` | Print version |

## Development

```bash
make build      # build to ./bin/spk
make install    # copy to /usr/local/bin
make test       # run tests
```

## Release

Push a tag to trigger automated release + Homebrew formula update:

```bash
git tag v0.2.0
git push --tags
```

GoReleaser cross-compiles for macOS (Intel + Apple Silicon) and Linux, creates a GitHub release, and auto-updates `Formula/spk.rb` in this repo.
