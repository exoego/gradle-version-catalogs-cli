![GitHub Repo stars](https://img.shields.io/github/stars/exoego/gradle-version-catalogs-cli?style=social)
![GitHub](https://img.shields.io/github/license/exoego/gradle-version-catalogs-cli)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/exoego/gradle-version-catalogs-cli)
![GitHub all releases](https://img.shields.io/github/downloads/exoego/gradle-version-catalogs-cli/total)
![GitHub CI Status](https://img.shields.io/github/actions/workflow/status/exoego/gradle-version-catalogs-cli/ci.yaml?branch=main&label=CI)
![GitHub Release Status](https://img.shields.io/github/v/release/exoego/gradle-version-catalogs-cli)
[![codecov](https://codecov.io/gh/exoego/gradle-version-catalogs-cli/graph/badge.svg?token=TP2GijRLFG)](https://codecov.io/gh/exoego/gradle-version-catalogs-cli)

# gradle-version-catalogs-cli

Generate [libs.version.toml (Version Catalogs)](https://docs.gradle.org/current/userguide/version_catalogs.html) from existing Gradle build files.

## Usage

1. Download from https://github.com/exoego/gradle-version-catalogs-cli/releases
2. Go to the repo where gradle is used
3. Run the command below

```bash
./gradle-version-cataglos-cli generate
```

## Commands

### Generate

```bash
gradle-version-catalogs-cli generate [PATH]
```  

- Collects library versions in multiple build.gradle(.kts) and generates libs.versions.toml in `PATH/gradle`.
- If no `PATH` is provided, the current working directory is used.
- Some manual intervention may be required.

> [!WARNING]
> If libs.version.toml already exists, it will be overwritten.

## Development

```bash
go get
make test
make build
./gradle-version-catalogs-cli ../path/to/gradle/project
```