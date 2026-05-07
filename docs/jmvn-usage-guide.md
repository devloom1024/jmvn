# jmvn 使用说明

## 1. 什么是 jmvn

`jmvn` 是一个跨平台的 Maven CLI 包装器，核心功能是**按项目自动选择 JDK、Maven 版本、`settings.xml` 和本地仓库路径**。

它的工作方式与 IntelliJ IDEA 相同——不通过 `mvn`/`mvnw` 脚本，而是直接用目标 JDK 的 `java` 可执行文件启动 Maven launcher。这意味着你切换项目时无需手动修改 `JAVA_HOME` 或 `PATH`。

### 典型场景

| 场景 | 传统方式 | 用 jmvn |
|---|---|---|
| 不同项目用不同 JDK | 手动改 JAVA_HOME | 自动切换 |
| 多仓库/多用户隔离 | 手动指定 settings.xml | 自动配置 |
| 不同项目用不同 Maven 版本 | 维护多套安装 | 按版本号使用 |
| 构建前确认命令 | 改脚本再删除 | `--dry-run` 预览 |

---

## 2. 安装

### 2.1 源码构建

```bash
# 克隆仓库
git clone <repo-url> jmvn
cd jmvn

# 构建当前平台
make build
# 输出：bin/jmvn

# 构建所有平台
make build-all
# 输出：
#   bin/jmvn-windows-amd64.exe
#   bin/jmvn-linux-amd64
#   bin/jmvn-darwin-amd64
#   bin/jmvn-darwin-arm64
```

将 `bin/jmvn`（或 Windows 下的 `bin/jmvn.exe`）放到 `PATH` 中的某个目录即可使用。

### 2.2 运行测试

```bash
make test
```

---

## 3. 初始配置

`jmvn` 使用两种配置文件，均采用 TOML 格式。

### 3.1 初始化——交互式方式（推荐）

```bash
# 初始化全局配置（所有项目的默认设置）
jmvn init --global

# 在项目目录下初始化项目配置
cd /path/to/your/project
jmvn init
```

`jmvn init` 会逐行提示你输入：
- 项目使用的 JDK 版本号
- Maven 版本号
- settings.xml 路径
- 本地仓库路径

其中 settings 和 local-repo 留空则使用全局默认值。

### 3.2 手动创建配置文件

#### 全局配置 `~/.jmvn/config.toml`

此文件定义了你本机所有可用的 JDK 和 Maven 安装，以及全局默认值：

```toml
[defaults]
# 默认 JDK 版本号，对应下面 [jdks] 中的 key
jdk = "17"
# Maven 安装路径（注意是 maven_home，不是 maven 版本号）
maven_home = "D:/tools/apache-maven-3.9.6"
# 全局 settings.xml
settings = "C:/Users/you/.m2/settings.xml"
# 全局本地仓库
local_repo = "C:/Users/you/.m2/repository"

[jdks]
# 版本号 = 安装路径
"8"  = "D:/jdks/jdk-8u402"
"11" = "D:/jdks/jdk-11.0.22"
"17" = "D:/jdks/jdk-17.0.10"
"21" = "D:/jdks/jdk-21.0.2"

[mavens]
# 版本号 = 安装路径
"3.6" = "D:/tools/apache-maven-3.6.3"
"3.9" = "D:/tools/apache-maven-3.9.6"
```

> **注意**：`[defaults]` 中的 `jdk` 和 `maven_home` 字段类型不同——`jdk` 是版本号（字符串），`maven_home` 是完整路径。

#### 项目配置 `.jmvn.toml`

放在项目根目录，仅配置当前项目所需：

```toml
# JDK 版本号（对应全局配置 [jdks] 中的 key）
jdk = "17"

# Maven 版本号（对应全局配置 [mavens] 中的 key）
maven = "3.9"

# settings.xml 路径（支持相对路径和 ~ 展开）
settings = "./maven/settings.xml"

# 本地仓库路径（支持相对路径和 ~ 展开）
local_repo = "./.m2/repository"
```

---

## 4. 命令参考

### 4.1 根命令——运行 Maven

```bash
# 基础用法：所有参数透传给 Maven
jmvn clean install
jmvn compile
jmvn test -Dtest=MyTest
jmvn package -DskipTests

# 临时覆盖配置
jmvn --jdk 11 clean test          # 用 JDK 11 执行
jmvn --maven 3.6 clean install    # 用 Maven 3.6 执行
jmvn --settings ~/another-settings.xml package
jmvn --local-repo /tmp/m2-repo verify
```

### 4.2 预览命令（dry-run）

```bash
# 仅打印最终要执行的 Java 命令，不实际运行
jmvn --dry-run clean install

# 结合 --verbose 查看更多解析细节
jmvn --verbose --dry-run clean install
```

输出示例（dry-run 模式）：

```
JMvn > [sudo] /path/to/jdk-17/bin/java \
  -classpath /path/to/maven-3.9.6/boot/plexus-classworlds-2.0-SNAPSHOT.jar \
  -Dclassworlds.conf=/path/to/maven-3.9.6/bin/m2.conf \
  -Dmaven.home=/path/to/maven-3.9.6 \
  -Dmaven.multiModuleProjectDirectory=/path/to/project \
  org.codehaus.plexus.classworlds.launcher.Launcher \
  clean install
```

> **安全提示**：`--dry-run` 模式不会执行任何命令，且前缀为 `[sudo]` 仅代表"预览模式"，并非真正的 `sudo` 提权。

### 4.3 version——查看版本

```bash
jmvn version
# 输出：jmvn v1.0.0 (go1.23.0, windows/amd64)
```

### 4.4 list——查看已注册的工具链

```bash
jmvn list
```

输出示例：

```
=== JDKs ===
  8  ->  D:/jdks/jdk-8u402             ✓
  11  ->  D:/jdks/jdk-11.0.22          ✓
  17  ->  D:/jdks/jdk-17.0.10          ✓
  21  ->  D:/jdks/jdk-21.0.2           ✓

=== Mavens ===
  3.6  ->  D:/tools/apache-maven-3.6.3  ✓
  3.9  ->  D:/tools/apache-maven-3.9.6  ✓
```

`✓` 表示路径存在且可访问，`✗` 表示路径不存在或不可访问。

### 4.5 info——查看当前项目解析结果

```bash
jmvn info
```

输出示例：

```
   Project Dir   C:/projects/my-app
   Java          D:/jdks/jdk-17.0.10/bin/java.exe  (project)
   Maven Home    D:/tools/apache-maven-3.9.6        (global)
   Settings      C:/projects/my-app/maven/settings.xml  (project)
   Local Repo    C:/Users/you/.m2/repository         (global)
   Config Files
     global      C:/Users/you/.jmvn/config.toml      found
     project     C:/projects/my-app/.jmvn.toml       found
```

括号中的 `(project)`/`(global)`/`(cli)`/`(env)` 表示该值的来源。

### 4.6 init——交互式初始化

```bash
# 初始化项目配置
cd /path/to/project
jmvn init

# 初始化全局配置
jmvn init --global
```

### 4.7 帮助

```bash
jmvn --help
jmvn init --help
```

---

## 5. 配置合并与优先级

当同时存在多种配置来源时，按以下优先级合并（高到低）：

```
1. CLI 参数     (--jdk, --maven, --settings, --local-repo)
      ↓
2. 项目配置     (.jmvn.toml)
      ↓
3. 全局默认值   (~/.jmvn/config.toml [defaults])
      ↓
4. 环境变量     (JAVA_HOME, MAVEN_HOME, M2_HOME)
```

**JDK 版本号解析流程**（仅当项目配置和 CLI 都未指定时自动触发）：

```
.java-version 文件 → pom.xml 属性 → .mvn/jdk.config 文件
```

- `.java-version`：文件内容如 `17`、`1.8`（会自动转为 `8`）
- `pom.xml`：按优先级检测 `<maven.compiler.release>` → `<maven.compiler.source>` → `<java.version>` → maven-compiler-plugin 配置
- `.mvn/jdk.config`：从文件中正则提取版本号

---

## 6. 实际工作流示例

### 6.1 新手入门流程

```bash
# 第一步：安装 jmvn
make build
cp bin/jmvn /usr/local/bin/

# 第二步：初始化全局配置（定义你的所有 JDK 和 Maven 安装）
jmvn init --global

# 第三步：在项目中初始化
cd ~/projects/my-spring-boot-app
jmvn init
# 交互输入项目所需的 JDK 版本和 Maven 版本

# 第四步：开始使用
jmvn compile
jmvn test
jmvn package
```

### 6.2 多 JDK 项目切换

```bash
# 在 Java 8 项目中
cd ~/projects/legacy-app
# .jmvn.toml 中 jdk = "8"
jmvn clean install    # 自动使用 JDK 8

# 切到 Java 17 项目
cd ~/projects/new-service
# .jmvn.toml 中 jdk = "17"
jmvn clean install    # 自动使用 JDK 17
```

无需手动修改 `JAVA_HOME`。

---

## 7. 工作原理

`jmvn` 并不调用 `mvn` 脚本。它先合并所有配置源，解析出目标 JDK 和 Maven 安装目录，然后直接通过 Java 启动 Maven launcher：

```
{JAVA_HOME}/bin/java
  [.mvn/jvm.config 中的 JVM 参数（每行一个参数）]
  -classpath {MAVEN_HOME}/boot/plexus-classworlds-*.jar
  -Dclassworlds.conf={MAVEN_HOME}/bin/m2.conf
  -Dmaven.home={MAVEN_HOME}
  -Dmaven.multiModuleProjectDirectory={PROJECT_DIR}
  [-Dmaven.repo.local={LOCAL_REPO}]
  [--settings {SETTINGS}]
  org.codehaus.plexus.classworlds.launcher.Launcher
  {maven 参数...}
```

你可以通过 `--dry-run` 随时查看实际会执行的命令。

---

## 8. 常见问题

### Q: `--jdk` 参数应该传什么值？
传的是全局配置 `[jdks]` 中定义的版本 **key**，例如 `--jdk 11`。它对应 `~/.jmvn/config.toml` 中 `[jdks]"11"` 的值。

### Q: 报错 "jdk 17 is not registered" 怎么办？
表示全局配置中 `[jdks]` 没有 `"17"` 这个 key。用 `jmvn init --global` 添加，或手动编辑 `~/.jmvn/config.toml`。

### Q: 可以不配置全局配置，只使用项目配置吗？
可以。但项目配置中的 `jdk` 和 `maven` 字段需要是版本号。如果项目配置中不指定这些版本号，jmvn 会回退到环境变量 `JAVA_HOME`/`MAVEN_HOME`。

### Q: `.jmvn.toml` 中的路径支持相对路径吗？
支持。相对路径相对于 `.jmvn.toml` 所在的目录解析。同时也支持 `~` 展开为用户主目录。

### Q: 支持 Maven Wrapper (mvnw) 吗？
jmvn 的设计初衷就是代替 `mvn`/`mvnw` 脚本，直接通过 Java 启动 Maven launcher。如果需要使用项目自带的 `.mvn/wrapper`，可以考虑用 `--dry-run` 确认命令后再决定。

### Q: 可以在 CI/CD 中使用吗？
可以。在 CI 环境中，可以通过环境变量或 CLI 参数来指定所有配置，无需创建配置文件：

```bash
export JAVA_HOME=/path/to/jdk-17
export MAVEN_HOME=/path/to/maven-3.9.6
jmvn clean verify
```
