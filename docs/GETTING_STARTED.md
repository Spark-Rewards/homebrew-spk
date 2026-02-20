# Getting started with spk (super simple guide)

This doc explains **what spk is** and **how to use it** in plain terms. No prior multi-repo or CLI experience required.

---

## What is spk?

**spk** is a small command-line tool that helps you work with **multiple Git repos in one place** and keeps things like **secrets and build order** under control.

Think of it like this:

- We have lots of separate repos (AppModel, AppAPI, MobileApp, BusinessAPI, etc.).
- You could clone each repo into random folders and run each app’s build by hand—but that gets messy and easy to mess up.
- **spk** gives you one “workspace” folder where all those repos live together, and one set of commands to **clone**, **sync**, **run builds**, and **share environment variables** (like API keys and config).

So: **spk = “workspace manager” for our multi-repo setup.**

---

## Why do we use it?

1. **One place for everything** – All the repos you need sit in one workspace folder (e.g. `~/SparkRewards`).
2. **Shared environment** – One `.env` file at the workspace root; spk can fill it from AWS (SSM) and optionally symlink it into repos so everyone uses the same config.
3. **Smart builds** – When you run `spk run build` in a repo that depends on another (e.g. AppAPI depends on AppModel), spk can build dependencies first and link local packages so you’re not constantly publishing to npm.
4. **Sync and credentials** – `spk sync` updates all repos (git fetch + rebase) and refreshes the workspace `.env` from AWS, and can trigger AWS SSO login if needed.

So: **spk keeps multi-repo development consistent and less error-prone.**

---

## Install

You need **Homebrew** (macOS or Linux).

```bash
brew tap Spark-Rewards/spk
brew install spk
```

Check it’s there:

```bash
spk --help
```

---

## How to use it (step by step)

### 1. Create a workspace

A **workspace** is just a folder that spk “owns”: it will put a `.spk` directory inside it (with `workspace.json`) and later a `.env` file there.

```bash
spk create workspace ~/SparkRewards
```

- Use any path you want (e.g. `./my-dev`, `~/Projects/SparkRewards`).
- If the folder doesn’t exist, spk creates it.
- Don’t run this inside an existing repo; create the workspace first, then add repos into it.

**What happened:** You now have a folder (e.g. `~/SparkRewards`) with a `.spk/workspace.json` file. That’s your workspace root.

---

### 2. Go into the workspace

```bash
cd ~/SparkRewards
```

All spk commands that “need a workspace” look for `.spk/workspace.json` in the current directory or any parent. So from now on, run spk from inside this folder (or inside any repo under it).

---

### 3. Add repos with `spk use`

“Use” here means “add this repo to my workspace” (clone it if needed and register it in `workspace.json`).

```bash
spk use AppModel
spk use AppAPI
spk use MobileApp
```

- Repo name only (e.g. `AppModel`) → spk assumes **Spark-Rewards** org and clones `Spark-Rewards/AppModel`.
- You can use `org/repo` or a full Git URL if needed.

Each repo ends up as a subfolder: e.g. `~/SparkRewards/AppModel`, `~/SparkRewards/AppAPI`.

---

### 4. Sync repos and environment: `spk sync`

This does two things:

1. **Git:** For each repo in the workspace, it does fetch + rebase (so you’re up to date with the remote).
2. **Environment:** It pulls values from AWS (SSM) and writes them into the workspace `.env` file. If your AWS session is expired, it can prompt you to log in (SSO).

```bash
spk sync
```

- **First time or after long idle:** You may need to log in to AWS. Run `spk login` if you want to do that separately, or let `spk sync` prompt you.
- **Only sync one repo:** `spk sync AppAPI`
- **Skip refreshing .env:** `spk sync --no-env`
- **Use a different env (e.g. prod):** `spk sync --env prod`

So: **sync = “update all my repos and refresh my shared .env from AWS.”**

---

### 5. Run scripts with `spk run`

When you’re **inside a repo directory** (e.g. `~/SparkRewards/AppAPI`), you can run scripts through spk so it can inject the workspace `.env` and handle dependencies.

```bash
cd ~/SparkRewards/AppAPI
spk run build
```

- spk figures out the project type (Node/npm, Gradle, Go, Make) and runs the right command (e.g. `npm run build`, `./gradlew build`).
- **List scripts:** Run `spk run` with no arguments to see what’s available in the current repo.
- **Build with dependencies:** If this repo depends on another (e.g. AppModel), use:
  ```bash
  spk run build -r
  ```
  The `-r` (recursive) flag builds dependency repos first, then the current one, and can link local packages so you don’t need to publish to npm.

So: **spk run = “run a script in this repo, with workspace env and optional dependency building.”**

---

## Command cheat sheet (plain English)

| Command | What it does |
|--------|----------------|
| `spk create workspace <path>` | Create a new workspace folder; spk will put `.spk/workspace.json` there. |
| `spk use <repo>` | Clone the repo into the workspace (if needed) and add it to the manifest. e.g. `spk use AppAPI`. |
| `spk sync` | Update all workspace repos (fetch + rebase) and refresh the shared `.env` from AWS. |
| `spk sync <repo>` | Same as above but only for that repo. |
| `spk run` | (Inside a repo.) List available scripts (e.g. build, test, start). |
| `spk run <script>` | (Inside a repo.) Run that script with workspace env; e.g. `spk run build`, `spk run test`. |
| `spk run build -r` | Build this repo after building its dependencies; uses local linked packages when possible. |
| `spk info` | Show workspace name, path, repos, their branches and status (clean/dirty). Aliases: `spk status`, `spk ws`. |
| `spk env` | Show current workspace environment variables (from the root `.env`). |
| `spk env set KEY=value` | Set (or overwrite) a variable in the workspace `.env`. |
| `spk env link` | Symlink each repo’s `.env` to the workspace `.env` so all repos share the same env. |
| `spk login` | Log in to AWS SSO (e.g. when `spk sync` says your session expired). |
| `spk remove <repo>` | Remove the repo from the workspace manifest only; does **not** delete the repo folder. |

---

## Where things live

- **Workspace root:** The folder you created with `spk create workspace <path>` (e.g. `~/SparkRewards`).
- **`.spk/workspace.json`** – List of repos and workspace settings (AWS profile, region, etc.). Don’t edit by hand unless you know what you’re doing.
- **`.env`** – At the workspace root. Filled by `spk sync` from AWS (and optionally by `spk env set`). Used when you run `spk run` so scripts see the same config.
- **Repos** – Each one is a normal Git repo in a subfolder (e.g. `AppAPI`, `AppModel`).

---

## Tips for newbies

1. **Always start from the workspace root** when you run `spk use` or `spk sync`. For `spk run`, you must be inside a repo directory.
2. **“Not inside a spk workspace”** – You’re not in the workspace folder (or a subfolder of it). `cd` to your workspace root and try again.
3. **First time AWS** – If `spk sync` or `spk login` fails, you may need to run `aws configure sso` once; spk will show instructions if SSO isn’t set up.
4. **Build failures** – If a build needs another repo (e.g. AppAPI needs AppModel), try `spk run build -r` from the repo that’s failing.
5. **VS Code** – spk can generate a multi-root workspace file (e.g. `SparkRewards.code-workspace`) so you can open all repos in one VS Code window. It’s created/updated when you create the workspace and when you add repos or run `spk sync`.

---

## Quick start (copy-paste)

```bash
# Install
brew tap Spark-Rewards/spk
brew install spk

# Create workspace and add repos
spk create workspace ~/SparkRewards
cd ~/SparkRewards
spk use AppModel
spk use AppAPI
spk use MobileApp

# Sync everything and refresh .env (will prompt for AWS if needed)
spk sync

# Build a repo (from inside that repo)
cd AppAPI
spk run build
# Or build with dependencies first:
spk run build -r
```

That’s it. For more, run `spk --help` and `spk <command> --help`.
