# jmvn 技术方案

> 一个跨平台的 Maven JDK 自动切换 CLI 工具，采用 IDEA 方式直接调用 Java 启动 Maven，实现项目级 JDK、Maven 版本、settings 和本地仓库的隔离管理。

## 1. 项目概述

### 1.1 解决的问题

在 Windows/macOS/Linux 上进行 Java 开发时，不同项目使用不同的 JDK 版本，但 `mvn` 命令始终使用 `JAVA_HOME` 指定的 JDK。手动切换 `JAVA_HOME` 既繁琐又容易出错。同时，不同项目可能需要不同的 Maven 版本、`settings.xml` 和本地仓库路径。

### 1.2 核心方案

**模仿 IntelliJ IDEA 的实现方式**：不调用 `mvn` / `mvn.cmd` 脚本，而是直接用指定 JDK 的 `java` 可执行文件启动 Maven 的 classworlds launcher。

```
jmvn clean install
  → 检测项目 JDK/Maven/Settings/Repo 配置
  → 直接执行: /path/to/jdk/bin/java -cp plexus-classworlds.jar ... Launcher clean install
```

### 1.3 技术选型

| 组件 | 选择 | 说明 |
|------|------|------|
| 语言 | Go 1.21+ | 编译为单文件可执行文件，零依赖，启动快 |
| CLI 框架 | spf13/cobra | 子命令支持，kubectl/docker/gh 都在用 |
| 配置解析 | BurntSushi/toml | Go 社区最经典的 TOML 库 |
| 终端彩色输出 | fatih/color | 跨平台（含 Windows）彩色终端输出 |
| XML 解析 | encoding/xml（标准库） | 解析 pom.xml 用 |
| 进程管理 | os/exec（标准库） | 启动 Java 子进程 |
| 文件查找 | filepath（标准库） | 跨平台路径处理 |

### 1.4 跨平台编译目标

```bash
GOOS=windows GOARCH=amd64  → jmvn.exe
GOOS=darwin  GOARCH=amd64  → jmvn (macOS Intel)
GOOS=darwin  GOARCH=arm64  → jmvn (macOS Apple Silicon)
GOOS=linux   GOARCH=amd64  → jmvn (Linux)
```

---

## 2. 配置文件设计

### 2.1 全局配置 `~/.jmvn/config.toml`

```toml
# ============================================================
# jmvn 全局配置
# ============================================================

# 默认配置（所有项目的兜底值）
[defaults]
jdk = "17"                                    # 默认 JDK 版本
maven_home = "C:\\apache-maven-3.9.6"         # 默认 Maven 安装目录
settings = "~/.m2/settings.xml"               # 默认 Maven settings.xml
local_repo = "~/.m2/repository"               # 默认本地仓库路径

# JDK 版本 → 安装路径映射
[jdks]
8  = "C:\\Program Files\\Java\\jdk1.8.0_361"
11 = "C:\\Program Files\\Java\\jdk-11.0.20"
17 = "C:\\Program Files\\Java\\jdk-17.0.8"
21 = "C:\\Program Files\\Java\\jdk-21"

# Maven 版本 → 安装路径映射
[mavens]
"3.6" = "C:\\apache-maven-3.6.3"
"3.9" = "C:\\apache-maven-3.9.6"
"4.0" = "C:\\apache-maven-4.0.0"
```

**macOS/Linux 示例：**

```toml
[defaults]
jdk = "17"
maven_home = "/opt/apache-maven-3.9.6"
settings = "~/.m2/settings.xml"
local_repo = "~/.m2/repository"

[jdks]
8  = "/usr/lib/jvm/java-8-openjdk"
11 = "/usr/lib/jvm/java-11-openjdk"
17 = "/usr/lib/jvm/java-17-openjdk"
21 = "/usr/lib/jvm/java-21-openjdk"

[mavens]
"3.6" = "/opt/apache-maven-3.6.3"
"3.9" = "/opt/apache-maven-3.9.6"
"4.0" = "/opt/apache-maven-4.0.0"
```

### 2.2 项目级配置 `.jmvn.toml`（项目根目录）

```toml
# 项目级配置（覆盖全局 defaults）
jdk = "11"                                     # 本项目使用 JDK 11
maven = "3.6"                                  # 本项目使用 Maven 3.6
settings = "./maven/settings-internal.xml"     # 项目专用 settings（支持相对路径）
local_repo = "D:\\maven-repos\\project-x"      # 项目专用本地仓库
```

所有字段均为可选，未指定的字段回退到全局 defaults。

### 2.3 配置解析优先级

每个配置项独立解析，从高到低：

```
1. 命令行参数        --jdk 11 / --maven 3.6 / --settings xxx / --local-repo xxx
2. 项目 .jmvn.toml   项目根目录下的配置文件
3. 全局 config.toml   ~/.jmvn/config.toml 中的 [defaults]
4. 环境变量兜底       JAVA_HOME / MAVEN_HOME（仅在以上都未指定时使用）
5. PATH 查找          从 PATH 中找 java / mvn（最终兜底）
```

### 2.4 Go 数据结构

```go
// GlobalConfig 全局配置文件结构
type GlobalConfig struct {
    Defaults DefaultsConfig       `toml:"defaults"`
    JDKs     map[string]string    `toml:"jdks"`
    Mavens   map[string]string    `toml:"mavens"`
}

type DefaultsConfig struct {
    JDK       string `toml:"jdk"`
    MavenHome string `toml:"maven_home"`
    Settings  string `toml:"settings"`
    LocalRepo string `toml:"local_repo"`
}

// ProjectConfig 项目级配置文件结构
type ProjectConfig struct {
    JDK       string `toml:"jdk"`
    Maven     string `toml:"maven"`
    Settings  string `toml:"settings"`
    LocalRepo string `toml:"local_repo"`
}

// ResolvedConfig 最终解析后的配置
type ResolvedConfig struct {
    JavaCmd   string // 最终的 java 可执行文件绝对路径
    MavenHome string // Maven 安装目录
    Settings  string // settings.xml 路径（可为空）
    LocalRepo string // 本地仓库路径（可为空）

    // 来源追踪（用于 jmvn info 显示）
    JavaCmdSource   string // "cli" / "project" / "global" / "env" / "path"
    MavenHomeSource string
    SettingsSource  string
    LocalRepoSource string
}
```

---

## 3. 项目 JDK 版本检测

当项目 `.jmvn.toml` 未指定 `jdk` 时，自动从以下来源检测（按优先级）：

### 3.1 检测顺序

| 优先级 | 来源 | 文件 | 示例内容 |
|--------|------|------|----------|
| 1 | `.java-version` | 项目根目录 | `17` |
| 2 | `pom.xml` 属性 | `maven.compiler.release` | `<maven.compiler.release>17</maven.compiler.release>` |
| 3 | `pom.xml` 属性 | `maven.compiler.source` | `<maven.compiler.source>11</maven.compiler.source>` |
| 4 | `pom.xml` 属性 | `java.version`（Spring Boot） | `<java.version>17</java.version>` |
| 5 | `pom.xml` 插件 | `maven-compiler-plugin` 的 `<release>` 或 `<source>` | `<release>17</release>` |
| 6 | `.mvn/jdk.config` | Maven 4 配置 | 解析其中指定的 JDK 路径/版本 |

### 3.2 pom.xml 解析策略

只做**浅层解析**，不做完整 Maven POM 继承解析：

```go
// PomProperties 从 pom.xml 提取的属性
type PomProperties struct {
    CompilerRelease string // maven.compiler.release
    CompilerSource  string // maven.compiler.source
    JavaVersion     string // java.version (Spring Boot convention)
    PluginRelease   string // maven-compiler-plugin <release>
    PluginSource    string // maven-compiler-plugin <source>
}
```

使用 `encoding/xml` 标准库 + 简单 XPath 式查找，不引入额外依赖。

---

## 4. Maven 启动原理

### 4.1 Maven 启动流程（等价于 mvn 脚本所做的事）

`mvn` 脚本的本质是构造以下 Java 命令：

```bash
$JAVA_HOME/bin/java \
  $MAVEN_OPTS \                                          # 用户 JVM 参数
  $JVM_CONFIG \                                          # .mvn/jvm.config 中的参数
  -classpath "$MAVEN_HOME/boot/plexus-classworlds-*.jar" \  # classworlds 启动器
  "-Dclassworlds.conf=$MAVEN_HOME/bin/m2.conf" \
  "-Dmaven.home=$MAVEN_HOME" \
  "-Dmaven.multiModuleProjectDirectory=$PROJECT_DIR" \
  org.codehaus.plexus.classworlds.launcher.Launcher \    # 启动器主类
  "$@"                                                   # Maven 参数透传
```

### 4.2 jmvn 构造的完整命令

```bash
"/path/to/jdk-17/bin/java" \
  $MAVEN_OPTS \
  $JVM_CONFIG \
  -classpath "/path/to/maven-3.9/boot/plexus-classworlds-2.7.0.jar" \
  "-Dclassworlds.conf=/path/to/maven-3.9/bin/m2.conf" \
  "-Dmaven.home=/path/to/maven-3.9" \
  "-Dmaven.multiModuleProjectDirectory=/path/to/project" \
  org.codehaus.plexus.classworlds.launcher.Launcher \
  --settings "/path/to/settings.xml" \                    # ← 配置的 settings
  "-Dmaven.repo.local=/path/to/local-repo" \              # ← 配置的 local repo
  clean install                                           # ← 用户参数
```

### 4.3 关键文件查找

```go
// findClassworldsJar 在 MAVEN_HOME/boot/ 下查找 plexus-classworlds-*.jar
func findClassworldsJar(mavenHome string) (string, error) {
    pattern := filepath.Join(mavenHome, "boot", "plexus-classworlds-*.jar")
    matches, err := filepath.Glob(pattern)
    // 返回第一个匹配
}

// findM2Conf 返回 MAVEN_HOME/bin/m2.conf 路径
func findM2Conf(mavenHome string) string {
    return filepath.Join(mavenHome, "bin", "m2.conf")
}

// findProjectBaseDir 向上查找包含 .mvn 目录的祖先目录
func findProjectBaseDir(startDir string) string {
    // 从 startDir 向上逐级查找 .mvn/ 目录
    // 找不到则返回 startDir（当前目录）
}
```

### 4.4 Maven 4 兼容处理

Maven 4 需要额外的 JVM 参数：

```go
func buildJvmArgs(javaVersion int, mavenVersion string) []string {
    args := []string{}
    // Maven 4.x 需要 --enable-native-access=ALL-UNNAMED
    if strings.HasPrefix(mavenVersion, "4.") {
        args = append(args, "--enable-native-access=ALL-UNNAMED")
    }
    return args
}
```

---

## 5. 项目结构

```
jmvn/
├── go.mod
├── go.sum
├── main.go                    # 入口
├── cmd/                       # CLI 命令定义（cobra）
│   ├── root.go                # 根命令：jmvn [maven-args...]（透传给 Maven）
│   ├── info.go                # jmvn info：显示当前配置解析结果
│   ├── init_cmd.go            # jmvn init：交互式初始化项目配置
│   ├── list.go                # jmvn list：列出所有已注册的 JDK 和 Maven
│   └── version.go             # jmvn version：显示 jmvn 自身版本
├── internal/
│   ├── config/
│   │   ├── global.go          # 全局配置加载 ~/.jmvn/config.toml
│   │   ├── project.go         # 项目配置加载 .jmvn.toml
│   │   └── resolve.go         # 配置合并、优先级解析
│   ├── detect/
│   │   ├── jdk.go             # JDK 版本自动检测（.java-version, pom.xml）
│   │   ├── pom.go             # pom.xml 解析
│   │   └── maven_home.go      # Maven 安装目录发现
│   ├── runner/
│   │   ├── builder.go         # 构建 Java 启动命令
│   │   └── executor.go        # 执行子进程，透传 stdin/stdout/stderr
│   └── util/
│       ├── path.go            # 路径工具（~ 展开、相对路径解析）
│       └── validate.go        # 验证 JDK/Maven 路径有效性
├── Makefile                   # 构建脚本
└── README.md
```

---

## 6. CLI 命令设计

### 6.1 主命令：`jmvn [maven-args...]`

作为 `mvn` 的透明代理，所有非 jmvn 参数直接透传给 Maven。

```bash
# 等价于 mvn clean install，但自动使用正确的 JDK
jmvn clean install

# 手动覆盖 JDK 版本
jmvn --jdk 11 clean install

# 手动覆盖 Maven 版本
jmvn --maven 3.6 clean package

# 手动覆盖 settings
jmvn --settings ~/alt-settings.xml clean install

# 手动覆盖 local repo
jmvn --local-repo /tmp/repo clean install

# 多个覆盖组合
jmvn --jdk 8 --maven 3.6 --settings ./settings.xml clean deploy

# 显示实际执行的命令但不执行（dry-run）
jmvn --dry-run clean install
```

**jmvn 自有参数**（必须放在 Maven 参数之前）：

| 参数 | 缩写 | 说明 |
|------|------|------|
| `--jdk <version>` | `-j` | 指定 JDK 版本 |
| `--maven <version>` | `-m` | 指定 Maven 版本 |
| `--settings <path>` | `-s` | 指定 settings.xml 路径 |
| `--local-repo <path>` | `-r` | 指定本地仓库路径 |
| `--dry-run` | `-n` | 只打印命令，不执行 |
| `--verbose` | `-v` | 显示详细配置解析过程 |

### 6.2 子命令

#### `jmvn info`

显示当前项目的完整配置解析结果：

```
$ jmvn info

  jmvn 配置解析
  ─────────────────────────────────────────────────────────────
  JDK         11  → C:\Program Files\Java\jdk-11.0.20       [.jmvn.toml]
  Maven       3.9 → C:\apache-maven-3.9.6                   [config.toml]
  Settings    ./maven/settings-internal.xml                  [.jmvn.toml]
  Local Repo  ~/.m2/repository                               [config.toml]
  Project Dir D:\projects\my-app
  ─────────────────────────────────────────────────────────────
  Config Files:
    Global:  C:\Users\xxx\.jmvn\config.toml     ✓ found
    Project: D:\projects\my-app\.jmvn.toml      ✓ found
```

#### `jmvn list`

列出所有已注册的 JDK 和 Maven 版本：

```
$ jmvn list

  已注册的 JDK:
    8   C:\Program Files\Java\jdk1.8.0_361     ✓
    11  C:\Program Files\Java\jdk-11.0.20      ✓
    17  C:\Program Files\Java\jdk-17.0.8       ✓
    21  C:\Program Files\Java\jdk-21           ✗ 路径不存在

  已注册的 Maven:
    3.6  C:\apache-maven-3.6.3                 ✓
    3.9  C:\apache-maven-3.9.6                 ✓
    4.0  C:\apache-maven-4.0.0                 ✗ 路径不存在
```

#### `jmvn init`

在当前项目目录下交互式生成 `.jmvn.toml`：

```
$ jmvn init

  初始化项目 jmvn 配置
  ? JDK 版本 (可选，留空跳过): 17
  ? Maven 版本 (可选，留空使用默认):
  ? settings.xml 路径 (可选，留空使用默认):
  ? 本地仓库路径 (可选，留空使用默认):

  ✓ 已生成 .jmvn.toml
```

#### `jmvn version`

```
$ jmvn version
jmvn v1.0.0 (go1.21, windows/amd64)
```

---

## 7. 核心流程

### 7.1 主流程伪代码

```go
func Run(cliArgs CLIArgs, mavenArgs []string) error {
    // 1. 加载全局配置
    globalCfg := config.LoadGlobal("~/.jmvn/config.toml")

    // 2. 查找项目根目录（向上查找 .jmvn.toml 或 pom.xml）
    projectDir := detect.FindProjectRoot(cwd)

    // 3. 加载项目配置
    projectCfg := config.LoadProject(projectDir)

    // 4. 如果项目配置中没有指定 JDK，尝试自动检测
    if projectCfg.JDK == "" && cliArgs.JDK == "" {
        projectCfg.JDK = detect.DetectJDKVersion(projectDir)
    }

    // 5. 合并配置（CLI > 项目 > 全局 > 环境变量 > PATH）
    resolved := config.Resolve(cliArgs, projectCfg, globalCfg)

    // 6. 验证路径有效性
    if err := validate(resolved); err != nil {
        return err
    }

    // 7. 构建 Java 启动命令
    javaCmd := runner.BuildCommand(resolved, mavenArgs)

    // 8. 如果 dry-run，打印命令并退出
    if cliArgs.DryRun {
        printCommand(javaCmd)
        return nil
    }

    // 9. 打印简要信息
    printBanner(resolved)

    // 10. 执行，透传 stdin/stdout/stderr，返回退出码
    return runner.Exec(javaCmd)
}
```

### 7.2 命令构建详情

```go
func BuildCommand(cfg ResolvedConfig, mavenArgs []string) *exec.Cmd {
    args := []string{}

    // JVM 参数
    if mavenOpts := os.Getenv("MAVEN_OPTS"); mavenOpts != "" {
        args = append(args, splitArgs(mavenOpts)...)
    }

    // .mvn/jvm.config 中的参数
    if jvmConfig := readJvmConfig(cfg.ProjectDir); jvmConfig != "" {
        args = append(args, splitArgs(jvmConfig)...)
    }

    // Maven 4 额外参数
    if isMaven4(cfg.MavenHome) {
        args = append(args, "--enable-native-access=ALL-UNNAMED")
    }

    // classpath
    classworldsJar := findClassworldsJar(cfg.MavenHome)
    args = append(args, "-classpath", classworldsJar)

    // 系统属性
    args = append(args,
        fmt.Sprintf("-Dclassworlds.conf=%s", filepath.Join(cfg.MavenHome, "bin", "m2.conf")),
        fmt.Sprintf("-Dmaven.home=%s", cfg.MavenHome),
        fmt.Sprintf("-Dmaven.multiModuleProjectDirectory=%s", cfg.ProjectDir),
    )

    // Maven 4 额外系统属性
    if isMaven4(cfg.MavenHome) {
        args = append(args,
            fmt.Sprintf("-Dlibrary.jline.path=%s", filepath.Join(cfg.MavenHome, "lib", "jline-native")),
        )
    }

    // launcher 主类
    args = append(args, "org.codehaus.plexus.classworlds.launcher.Launcher")

    // settings（通过 Maven 原生参数传递）
    if cfg.Settings != "" {
        args = append(args, "--settings", cfg.Settings)
    }

    // local repo（通过 Maven 原生参数传递）
    if cfg.LocalRepo != "" {
        args = append(args, fmt.Sprintf("-Dmaven.repo.local=%s", cfg.LocalRepo))
    }

    // 用户的 Maven 参数透传
    args = append(args, mavenArgs...)

    cmd := exec.Command(cfg.JavaCmd, args...)
    cmd.Dir = cfg.ProjectDir
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    // 不设置 cmd.Env，继承当前环境，不污染 JAVA_HOME
    return cmd
}
```

### 7.3 进程执行与信号处理

```go
func Exec(cmd *exec.Cmd) error {
    // 启动子进程
    if err := cmd.Start(); err != nil {
        return err
    }

    // 转发信号（Ctrl+C 等）到子进程
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
    go func() {
        for sig := range sigCh {
            cmd.Process.Signal(sig)
        }
    }()

    // 等待子进程结束，使用其退出码
    if err := cmd.Wait(); err != nil {
        if exitErr, ok := err.(*exec.ExitError); ok {
            os.Exit(exitErr.ExitCode())
        }
        return err
    }
    return nil
}
```

---

## 8. JDK 版本自动检测

### 8.1 `.java-version` 文件

```go
func detectFromJavaVersion(projectDir string) string {
    path := filepath.Join(projectDir, ".java-version")
    content, err := os.ReadFile(path)
    if err != nil {
        return ""
    }
    // 内容示例: "17" 或 "17.0.8" 或 "1.8"
    version := strings.TrimSpace(string(content))
    return normalizeVersion(version) // "1.8" → "8", "17.0.8" → "17"
}
```

### 8.2 pom.xml 属性

```go
func detectFromPom(projectDir string) string {
    pomPath := filepath.Join(projectDir, "pom.xml")
    // 解析 <properties> 中的:
    //   maven.compiler.release  (优先级最高)
    //   maven.compiler.source
    //   java.version           (Spring Boot 约定)
    // 解析 maven-compiler-plugin 配置中的:
    //   <release>
    //   <source>
}
```

### 8.3 版本号归一化

```go
// normalizeVersion 将各种格式的版本号归一化为主版本号
// "1.8"     → "8"
// "1.8.0"   → "8"
// "11.0.20" → "11"
// "17"      → "17"
func normalizeVersion(v string) string {
    if strings.HasPrefix(v, "1.") {
        parts := strings.Split(v, ".")
        if len(parts) >= 2 {
            return parts[1]
        }
    }
    parts := strings.Split(v, ".")
    return parts[0]
}
```

---

## 9. Maven 安装目录发现

当配置中未指定 `maven_home` 时，自动发现 Maven 安装位置：

```go
func discoverMavenHome() (string, error) {
    // 1. 环境变量 MAVEN_HOME 或 M2_HOME
    if home := os.Getenv("MAVEN_HOME"); home != "" {
        return home, nil
    }
    if home := os.Getenv("M2_HOME"); home != "" {
        return home, nil
    }

    // 2. 从 PATH 中找到 mvn，反推 MAVEN_HOME
    mvnPath, err := exec.LookPath("mvn")
    if err == nil {
        // mvn 在 MAVEN_HOME/bin/mvn，所以向上两级
        binDir := filepath.Dir(mvnPath)
        // 处理 symlink
        realPath, _ := filepath.EvalSymlinks(binDir)
        return filepath.Dir(realPath), nil
    }

    return "", fmt.Errorf("未找到 Maven 安装，请在 ~/.jmvn/config.toml 中配置 maven_home")
}
```

---

## 10. 错误处理

### 10.1 错误场景与提示

| 错误场景 | 提示信息 |
|----------|----------|
| 全局配置不存在 | `~/.jmvn/config.toml 不存在，运行 jmvn init --global 初始化` |
| JDK 版本未注册 | `JDK 17 未在 config.toml 中注册，已注册的版本: 8, 11` |
| JDK 路径不存在 | `JDK 17 路径不存在: C:\...\jdk-17（请检查 config.toml [jdks] 配置）` |
| java 可执行文件不存在 | `java 可执行文件不存在: C:\...\jdk-17\bin\java.exe` |
| Maven 目录不存在 | `Maven 目录不存在: C:\apache-maven-3.9.6` |
| plexus-classworlds JAR 找不到 | `在 C:\...\boot\ 下未找到 plexus-classworlds-*.jar，Maven 安装可能损坏` |
| settings.xml 不存在 | `settings.xml 不存在: ./maven/settings.xml` |
| 项目无 pom.xml | `当前目录下未找到 pom.xml` |
| 无法检测 JDK 版本 | `未能自动检测 JDK 版本，使用默认版本 17（来自全局配置）` |

### 10.2 错误退出码

| 退出码 | 含义 |
|--------|------|
| 0 | 成功（Maven 成功执行） |
| 1 | jmvn 自身错误（配置错误、路径不存在等） |
| N (>1) | Maven 的退出码（透传） |

---

## 11. 构建与发布

### 11.1 Makefile

```makefile
VERSION := 1.0.0
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"

.PHONY: build build-all clean

build:
	go build $(LDFLAGS) -o bin/jmvn .

build-all:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/jmvn-windows-amd64.exe .
	GOOS=darwin  GOARCH=amd64 go build $(LDFLAGS) -o bin/jmvn-darwin-amd64 .
	GOOS=darwin  GOARCH=arm64 go build $(LDFLAGS) -o bin/jmvn-darwin-arm64 .
	GOOS=linux   GOARCH=amd64 go build $(LDFLAGS) -o bin/jmvn-linux-amd64 .

clean:
	rm -rf bin/
```

### 11.2 GitHub Actions CI（可选）

使用 goreleaser 实现自动发布，在 tag push 时自动构建所有平台的二进制文件并发布到 GitHub Releases。

---

## 12. 使用示例

### 12.1 首次安装配置

```bash
# 1. 下载 jmvn 二进制文件，放入 PATH

# 2. 初始化全局配置
jmvn init --global
# 交互式填写 JDK/Maven 路径

# 3. 验证
jmvn list
```

### 12.2 项目使用

```bash
# 在项目目录下初始化
cd /path/to/my-project
jmvn init
# ? JDK 版本: 11
# ? Maven 版本: (留空使用默认)
# ? settings.xml: ./maven/settings.xml
# ✓ 已生成 .jmvn.toml

# 直接使用
jmvn clean install
# [jmvn] JDK 11 → C:\...\jdk-11 | Maven 3.9 → C:\...\mvn-3.9 | settings → ./maven/settings.xml
# [INFO] Scanning for projects...
# ...

# 查看配置
jmvn info

# 临时覆盖
jmvn --jdk 17 clean test
```

### 12.3 团队协作

`.jmvn.toml` 可提交到 Git 仓库，团队成员只需在全局配置中注册各自的 JDK/Maven 路径即可。

```gitignore
# .gitignore 中不需要忽略 .jmvn.toml
# 它只包含版本号，不包含机器相关路径
```

---

## 13. 实现计划

### Phase 1：核心功能（MVP）

1. 全局配置加载（`~/.jmvn/config.toml`）
2. 项目配置加载（`.jmvn.toml`）
3. 配置合并与优先级解析
4. Maven 启动命令构建（IDEA 方式）
5. 子进程执行与信号转发
6. `jmvn [maven-args]` 主命令

### Phase 2：辅助命令

7. `jmvn info` 命令
8. `jmvn list` 命令
9. `jmvn init` 命令（项目级 + 全局级）
10. `jmvn version` 命令

### Phase 3：智能检测

11. `.java-version` 检测
12. `pom.xml` JDK 版本检测
13. Maven 安装目录自动发现
14. `.mvn/jvm.config` 读取

### Phase 4：完善

15. `--dry-run` 支持
16. `--verbose` 详细日志
17. 跨平台测试与修复
18. README 文档
19. GitHub Actions 自动发布
