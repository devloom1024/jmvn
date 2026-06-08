# jmvn

`jmvn` 是一个跨平台的 Maven CLI 包装器，用来按项目自动选择 JDK、Maven、`settings.xml` 和本地仓库路径。

它不直接调用 `mvn` 脚本，而是像 IntelliJ IDEA 一样，直接使用目标 JDK 的 `java` 可执行文件启动 Maven launcher。

## 设计原则

**`jmvn` 是 `mvn` 的透明替代品。** 所有不以 `:` 开头的参数都原样透传给 Maven，零学习成本：

```bash
# 直接替换 mvn，一模一样
jmvn clean install
jmvn -pl flight-ticket-bdos-business -am -Dtest=SysSwtichConfigServiceImplTest test
jmvn -DskipTests package
```

以 `:` 开头的参数是 `jmvn` 的专属命令：

```bash
jmvn :init           # 初始化配置
jmvn :init --global  # 初始化全局配置
jmvn :info           # 查看当前解析的 JDK/Maven/settings
jmvn :list           # 列出已注册的工具链
jmvn :version        # jmvn 版本信息
jmvn :dry-run [args] # 预览将要执行的命令，不实际运行
jmvn :help           # 帮助信息
```

## 安装与构建

```bash
make build
make build-all   # 多平台
make test        # 运行测试
```

## 配置

配置文件体系（从高到低优先级）：

| 优先级 | 来源 | 说明 |
|--------|------|------|
| 1 | `JMVN_*` 环境变量 | 一次性覆盖 |
| 2 | `.jmvn.toml` | 项目级配置 |
| 3 | `~/.jmvn/config.toml` | 全局工具链注册 + 默认配置 |
| 4 | 自动检测 / `JAVA_HOME` | `.java-version` / `pom.xml` / 环境变量 |

### 全局配置 `~/.jmvn/config.toml`

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

### 项目配置 `.jmvn.toml`

```toml
jdk = "17"
maven = "3.9"
settings = "./maven/settings.xml"
local_repo = "./.m2/repository"
```

### 环境变量覆盖

```bash
# 一次性用 JDK 21 执行（不改配置文件）
JMVN_JDK=21 jmvn clean install

# 覆盖 settings / local-repo
JMVN_SETTINGS=/path/to/settings.xml jmvn clean test

# 直接指定 Maven 目录
JMVN_MAVEN_HOME=/path/to/maven jmvn clean install
```

## 工作原理

`jmvn` 会先合并环境变量、项目配置、全局配置和自动检测结果，再解析出目标 JDK 与 Maven 安装目录，最后构造类似下面的 Java 启动命令：

```bash
/path/to/jdk/bin/java \
  -classpath /path/to/maven/boot/plexus-classworlds-*.jar \
  -Dclassworlds.conf=/path/to/maven/bin/m2.conf \
  -Dmaven.home=/path/to/maven \
  -Dmaven.multiModuleProjectDirectory=/path/to/project \
  org.codehaus.plexus.classworlds.launcher.Launcher \
  clean install
```
