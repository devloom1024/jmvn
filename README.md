# jmvn

`jmvn` 是一个跨平台的 Maven CLI 包装器，用来按项目自动选择 JDK、Maven、`settings.xml` 和本地仓库路径。

它不直接调用 `mvn` 脚本，而是像 IntelliJ IDEA 一样，直接使用目标 JDK 的 `java` 可执行文件启动 Maven launcher。

## 当前能力

- 根命令透传 Maven 参数（等价于 `jmvn run`）
- 显式 `jmvn run` 子命令
- 支持 `--jdk`、`--maven`、`--settings`、`--local-repo`、`--dry-run`、`--verbose`
- 支持项目级 `.jmvn.toml` 和全局 `~/.jmvn/config.toml`
- 支持从 `.java-version` 和 `pom.xml` 检测项目 JDK
- 支持 `version`、`list`、`info`、`init`、`run`

## 安装与构建

本地构建：

```bash
make build
```

多平台构建：

```bash
make build-all
```

运行测试：

```bash
make test
```

## 全局配置示例

路径：`~/.jmvn/config.toml`

```toml
[defaults]
jdk = "17"
maven_home = "D:/mavens/apache-maven-3.9.6"
settings = "D:/users/demo/.m2/settings.xml"
local_repo = "D:/users/demo/.m2/repository"

[jdks]
"11" = "D:/jdks/jdk-11"
"17" = "D:/jdks/jdk-17"

[mavens]
"3.6" = "D:/mavens/apache-maven-3.6.3"
"3.9" = "D:/mavens/apache-maven-3.9.6"
```

## 项目配置示例

路径：`.jmvn.toml`

```toml
jdk = "17"
maven = "3.9"
settings = "./maven/settings.xml"
local_repo = "./.m2/repository"
```

## 常见命令

查看帮助：

```bash
jmvn --help
```

运行 Maven（两种等价方式）：

```bash
jmvn clean install
jmvn run clean install
```

查看版本：

```bash
jmvn version
```

查看当前解析结果：

```bash
jmvn info
```

查看已注册工具链：

```bash
jmvn list
```

初始化项目配置：

```bash
jmvn init
```

初始化全局配置：

```bash
jmvn init --global
```

打印实际启动命令但不执行：

```bash
jmvn --dry-run clean test
```

使用指定 JDK 查看解析结果：

```bash
jmvn info --jdk 8
```

## 工作原理

`jmvn` 会先合并 CLI 参数、项目配置、全局配置和环境变量，再解析出目标 JDK 与 Maven 安装目录，最后构造类似下面的 Java 启动命令：

```bash
/path/to/jdk/bin/java \
  -classpath /path/to/maven/boot/plexus-classworlds-*.jar \
  -Dclassworlds.conf=/path/to/maven/bin/m2.conf \
  -Dmaven.home=/path/to/maven \
  -Dmaven.multiModuleProjectDirectory=/path/to/project \
  org.codehaus.plexus.classworlds.launcher.Launcher \
  clean install
```

## 当前限制

- `init` 目前是最小可用交互实现
- `pom.xml` 检测目前覆盖的是常见属性场景
- `.mvn/jvm.config`、更完整的 Maven 4 兼容和发布细节仍在继续完善
