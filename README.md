# spk

Multi-repo workspace CLI for Spark Rewards development.

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
