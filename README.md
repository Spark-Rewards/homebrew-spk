# spk

Multi-repo workspace CLI for Spark Rewards development.

**New to spk?** See **[docs/GETTING_STARTED.md](docs/GETTING_STARTED.md)** for a guide (purpose, install, and how to use it).

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
