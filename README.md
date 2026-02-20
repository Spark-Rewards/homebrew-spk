# spark-cli

Multi-repo workspace CLI for Spark Rewards development.

**New to spark-cli?** See **[docs/GETTING_STARTED.md](docs/GETTING_STARTED.md)** for a guide (purpose, install, and how to use it).

## Install

```bash
brew install spark-rewards/spark-cli/spark-cli 
```

## Usage

```bash
spark-cli create workspace ~/SparkRewards
cd ~/SparkRewards
spark-cli use https://github.com/Spark-Rewards/BusinessAPI.git
spark-cli use [AppAPI](https://github.com/Spark-Rewards/BusinessModel.git
spark-cli sync
cd BusinessAPI && spark-cli run build
```

Run `spark-cli --help` for all commands.
