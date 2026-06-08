# Changelog

## [0.4.0](https://github.com/devloom1024/jmvn/compare/v0.3.0...v0.4.0) (2026-06-08)


### ⚠ BREAKING CHANGES

* CLI flags (--jdk, --dry-run, --verbose) replaced by environment variables (JMVN_JDK, etc.) and colon-prefix commands (:info, :dry-run, :init, :list, :version, :help). The `jmvn run` subcommand is removed; the root command is now the Maven entry point.

### Features

* redesign CLI with colon-prefix commands for transparent Maven passthrough ([d46857c](https://github.com/devloom1024/jmvn/commit/d46857cf01b633934ad53f0d35283a530e46856a))

## [Unreleased]

### BREAKING CHANGES

* 重构 CLI 设计：以 `:` 前缀区分 jmvn 命令，其余参数全透传 Maven
  * jmvn 命令：`:init`, `:info`, `:list`, `:version`, `:dry-run`, `:help`
  * CLI flag（`--jdk`, `--dry-run` 等）替换为环境变量（`JMVN_JDK` 等）
  * Maven 参数零冲突，`jmvn -pl module test` 开箱即用
  * 移除 `jmvn run` 子命令（根命令即 run）

## [0.3.0](https://github.com/devloom1024/jmvn/compare/v0.2.1...v0.3.0) (2026-05-19)


### Features

* add explicit 'jmvn run' subcommand ([#9](https://github.com/devloom1024/jmvn/issues/9)) ([60811dd](https://github.com/devloom1024/jmvn/commit/60811dd8fc34cc9925373ba22b39e41e422b1ac2))

## [0.2.1](https://github.com/devloom1024/jmvn/compare/v0.2.0...v0.2.1) (2026-05-12)


### Bug Fixes

* add checkout step to release job for gh CLI ([#7](https://github.com/devloom1024/jmvn/issues/7)) ([e385f5a](https://github.com/devloom1024/jmvn/commit/e385f5a5aa839d61a21bcdd903cea0ba249f7027))
* resolve Maven property placeholders & refactor build scripts ([#8](https://github.com/devloom1024/jmvn/issues/8)) ([224fdb6](https://github.com/devloom1024/jmvn/commit/224fdb66ce616dd7e1a4f94ad01759cee4493021))
* trigger release workflow on release published event ([#5](https://github.com/devloom1024/jmvn/issues/5)) ([7778d05](https://github.com/devloom1024/jmvn/commit/7778d05e7821f9b4fd240f633ecddafe34df5ff5))

## [0.2.0](https://github.com/devloom1024/jmvn/compare/v0.1.0...v0.2.0) (2026-05-08)


### Features

* MVP 版本——跨平台 Maven CLI 包装器 ([#1](https://github.com/devloom1024/jmvn/issues/1)) ([b05881b](https://github.com/devloom1024/jmvn/commit/b05881b30f58a256fb1bc247f21401906ccb0d23))
* 初始化工程 ([ae942a2](https://github.com/devloom1024/jmvn/commit/ae942a2ddaa2c8bbbb403767908b273df9fd5d7f))


### Bug Fixes

* add release-please config and checkout step ([#3](https://github.com/devloom1024/jmvn/issues/3)) ([0714b9d](https://github.com/devloom1024/jmvn/commit/0714b9d19ec4aeae4d8b48d6fa693c718ebe6e5f))
