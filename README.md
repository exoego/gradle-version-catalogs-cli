![GitHub Repo stars](https://img.shields.io/github/stars/exoego/gradle-version-catalogs-cli?style=social)
![GitHub](https://img.shields.io/github/license/exoego/gradle-version-catalogs-cli)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/exoego/gradle-version-catalogs-cli)
![GitHub all releases](https://img.shields.io/github/downloads/exoego/gradle-version-catalogs-cli/total)
![GitHub CI Status](https://img.shields.io/github/actions/workflow/status/exoego/gradle-version-catalogs-cli/ci.yaml?branch=main&label=CI)
![GitHub Release Status](https://img.shields.io/github/v/release/exoego/gradle-version-catalogs-cli)
[![codecov](https://codecov.io/gh/exoego/gradle-version-catalogs-cli/graph/badge.svg?token=TP2GijRLFG)](https://codecov.io/gh/exoego/gradle-version-catalogs-cli)

# gradle-version-catalogs-cli

- Generate [libs.version.toml (Version Catalogs)](https://docs.gradle.org/current/userguide/version_catalogs.html) from existing Gradle build files.
- And replace the dependencies in the build files build.gradle(.kts) with the version catalog.

Example)

```Kotlin
// Before 
val awsSdkVersion = "2.3.4"
implementation("software.amazon.awssdk:sts:$awsSdkVersion")
implementation("software.amazon.awssdk:dynamodb:$awsSdkVersion")
implementation("ch.qos.logback:logback-classic")
```

This will be rewritten to:

```Kotlin
// After
implementation(libs.software.amazon.awssdk.sts)
implementation(libs.software.amazon.awssdk.dynamodb)
implementation(libs.ch.qos.logback.logback.classic)
```

with the generated version catalogs `libs.versions.toml`: 

```toml
[versions]
awsSdkVersion = "2.3.4"

[dependencies]
software-amazon-awssdk-dynamodb = { group = "software.amazon.awssdk", name = "dynamodb", version.ref = "awsSdkVersion" }
software-amazon-awssdk-sts = { group = "software.amazon.awssdk", name = "sts", version.ref = "awsSdkVersion" }
ch-qos-logback-logback-classic = { group = "ch.qos.logback", name = "logback-classic", version = "1.5.18" }
```

> [!NOTE]
> In the above example, the latest version `1.5.18` is automatically selected for `logback-classic` since version must be specified in version catalog. If the command failed to decide latest version, it fallbacks to `FIXME`.
> This behavior is enabled by default and can be disabled via `--auto-latest=false` option.
> The autofilled version may be too new, so some manual intervention may be required.
  
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

## Development

```bash
go get
make test
make build
./gradle-version-catalogs-cli ../path/to/gradle/project
```
