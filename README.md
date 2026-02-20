# spk

Multi-repo workspace CLI for Spark Rewards development.

**New to spk?** See **[docs/GETTING_STARTED.md](docs/GETTING_STARTED.md)** for a guide (purpose, install, and how to use it).

**Release version:** The next release version is in **[VERSION](VERSION)**. CI uses it when tagging and releasing. To cut a minor or major release (e.g. `0.2.0`), update `VERSION` before pushing to `main`; otherwise each push releases that version and CI bumps the patch for next time.

## Install

```bash
brew tap Spark-Rewards/spk
brew install spk
```

## Usage

```bash
spk create workspace ~/SparkRewards
cd ~/SparkRewards
spk use AppModel
spk use AppAPI
spk sync
cd AppAPI && spk run build
```

Run `spk --help` for all commands.
