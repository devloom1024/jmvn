# jmvn Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 构建一个跨平台的 `jmvn` CLI，在执行 Maven 命令时按项目配置自动切换 JDK、Maven、`settings.xml` 与本地仓库路径。

**Architecture:** 采用 Cobra 组织 CLI 命令，`internal/config` 负责配置加载与优先级合并，`internal/detect` 负责项目根目录和 JDK/Maven 自动发现，`internal/runner` 负责拼装并执行 Java 启动命令。整体以 TDD 方式推进，优先实现可验证的最小闭环，再逐步补齐辅助命令与文档。

**Tech Stack:** Go 1.23、spf13/cobra、BurntSushi/toml、fatih/color、标准库 `os/exec`、`encoding/xml`、`filepath`

---

## Chunk 1: 项目骨架与配置解析

### Task 1: 初始化 CLI 与基础目录

**Files:**
- Modify: `go.mod`
- Modify: `main.go`
- Create: `cmd/root.go`
- Create: `cmd/root_test.go`
- Create: `internal/cli/options.go`

- [ ] **Step 1: 写根命令行为测试**

```go
func TestRootCommand_ParsesOwnFlagsAndLeavesMavenArgs(t *testing.T) {
    cmd := NewRootCmd()
    cmd.SetArgs([]string{"--jdk", "17", "--dry-run", "clean", "install"})

    opts, mvnArgs, err := executeForTest(cmd)

    require.NoError(t, err)
    require.Equal(t, "17", opts.JDK)
    require.True(t, opts.DryRun)
    require.Equal(t, []string{"clean", "install"}, mvnArgs)
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `go test ./cmd -run TestRootCommand_ParsesOwnFlagsAndLeavesMavenArgs -v`
Expected: FAIL，提示 `NewRootCmd` 或解析逻辑不存在。

- [ ] **Step 3: 最小实现根命令与参数结构**

```go
type Options struct {
    JDK       string
    Maven     string
    Settings  string
    LocalRepo string
    DryRun    bool
    Verbose   bool
}
```

在 `cmd/root.go` 中建立 `jmvn [maven-args...]` 根命令，预留执行入口，先完成参数解析与 Maven 参数透传。

- [ ] **Step 4: 再次运行测试确认通过**

Run: `go test ./cmd -run TestRootCommand_ParsesOwnFlagsAndLeavesMavenArgs -v`
Expected: PASS

- [ ] **Step 5: 提交这一小步**

```bash
git add go.mod main.go cmd/root.go cmd/root_test.go internal/cli/options.go
git commit -m "feat: bootstrap jmvn root command"
```

### Task 2: 实现路径工具与配置模型

**Files:**
- Create: `internal/config/types.go`
- Create: `internal/util/path.go`
- Create: `internal/util/path_test.go`

- [ ] **Step 1: 写路径展开与相对路径解析测试**

```go
func TestResolvePath_ExpandsHomeAndProjectRelative(t *testing.T) {
    got := ResolvePath("~/settings.xml", "D:/work/demo")
    require.Contains(t, got, "settings.xml")

    got = ResolvePath("./maven/settings.xml", "D:/work/demo")
    require.Equal(t, filepath.Clean("D:/work/demo/maven/settings.xml"), got)
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `go test ./internal/util -run TestResolvePath_ExpandsHomeAndProjectRelative -v`
Expected: FAIL，提示 `ResolvePath` 未实现。

- [ ] **Step 3: 实现配置结构与路径工具**

实现 `GlobalConfig`、`DefaultsConfig`、`ProjectConfig`、`ResolvedConfig`，并提供 `ExpandHome`、`ResolvePath`、`ResolveJavaBinary` 等工具函数。

- [ ] **Step 4: 运行测试确认通过**

Run: `go test ./internal/util -run TestResolvePath_ExpandsHomeAndProjectRelative -v`
Expected: PASS

- [ ] **Step 5: 提交这一小步**

```bash
git add internal/config/types.go internal/util/path.go internal/util/path_test.go
git commit -m "feat: add config models and path helpers"
```

### Task 3: 加载全局配置与项目配置

**Files:**
- Create: `internal/config/global.go`
- Create: `internal/config/project.go`
- Create: `internal/config/load_test.go`

- [ ] **Step 1: 写 TOML 配置加载测试**

```go
func TestLoadGlobal_ReadsDefaultsAndMaps(t *testing.T) {
    path := writeTempFile(t, `
[defaults]
jdk = "17"
maven_home = "/opt/maven"
[jdks]
"17" = "/opt/jdk17"
`)

    cfg, err := LoadGlobal(path)

    require.NoError(t, err)
    require.Equal(t, "17", cfg.Defaults.JDK)
    require.Equal(t, "/opt/jdk17", cfg.JDKs["17"])
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `go test ./internal/config -run TestLoadGlobal_ReadsDefaultsAndMaps -v`
Expected: FAIL，提示加载函数不存在或解析失败。

- [ ] **Step 3: 引入 TOML 依赖并实现加载逻辑**

在 `go.mod` 增加 `BurntSushi/toml`，实现：
- `LoadGlobal(path string) (GlobalConfig, error)`
- `LoadProject(path string) (ProjectConfig, error)`
- 缺失文件时返回可区分错误，供 CLI 决定是否降级处理

- [ ] **Step 4: 运行配置包测试**

Run: `go test ./internal/config -v`
Expected: PASS

- [ ] **Step 5: 提交这一小步**

```bash
git add go.mod internal/config/global.go internal/config/project.go internal/config/load_test.go
git commit -m "feat: load jmvn global and project config"
```

## Chunk 2: 发现逻辑、配置解析与 Maven 执行

### Task 4: 实现项目根目录与 JDK 版本检测

**Files:**
- Create: `internal/detect/project_root.go`
- Create: `internal/detect/jdk.go`
- Create: `internal/detect/pom.go`
- Create: `internal/detect/jdk_test.go`
- Test: `internal/detect/jdk_test.go`

- [ ] **Step 1: 写项目根目录与 `.java-version` 检测测试**

```go
func TestDetectJDKVersion_PrefersJavaVersionFile(t *testing.T) {
    projectDir := writeProjectTree(t, map[string]string{
        ".java-version": "17.0.8\n",
        "pom.xml": "<project></project>",
    })

    got := DetectJDKVersion(projectDir)

    require.Equal(t, "17", got)
}
```

- [ ] **Step 2: 写 `pom.xml` 属性检测测试**

```go
func TestDetectJDKVersion_FromPomCompilerRelease(t *testing.T) {
    projectDir := writeProjectTree(t, map[string]string{
        "pom.xml": `<project><properties><maven.compiler.release>21</maven.compiler.release></properties></project>`,
    })

    got := DetectJDKVersion(projectDir)

    require.Equal(t, "21", got)
}
```

- [ ] **Step 3: 运行测试确认失败**

Run: `go test ./internal/detect -run TestDetectJDKVersion -v`
Expected: FAIL

- [ ] **Step 4: 实现项目根目录发现、版本归一化与浅层 POM 解析**

实现：
- `FindProjectRoot(startDir string) string`
- `DetectJDKVersion(projectDir string) string`
- `ParsePomProperties(path string) (PomProperties, error)`
- `normalizeVersion(string) string`

- [ ] **Step 5: 运行发现模块测试**

Run: `go test ./internal/detect -v`
Expected: PASS

- [ ] **Step 6: 提交这一小步**

```bash
git add internal/detect/project_root.go internal/detect/jdk.go internal/detect/pom.go internal/detect/jdk_test.go
git commit -m "feat: detect project root and jdk version"
```

### Task 5: 实现配置优先级合并与校验

**Files:**
- Create: `internal/config/resolve.go`
- Create: `internal/util/validate.go`
- Create: `internal/config/resolve_test.go`
- Create: `internal/util/validate_test.go`

- [ ] **Step 1: 写配置优先级合并测试**

```go
func TestResolve_PrefersCLIThenProjectThenGlobal(t *testing.T) {
    resolved, err := Resolve(cliOpts, projectCfg, globalCfg, env)

    require.NoError(t, err)
    require.Equal(t, "cli", resolved.JavaCmdSource)
    require.Equal(t, "project", resolved.SettingsSource)
    require.Equal(t, "global", resolved.LocalRepoSource)
}
```

- [ ] **Step 2: 写路径有效性校验测试**

```go
func TestValidateResolvedConfig_ReturnsHelpfulErrorForMissingJava(t *testing.T) {
    err := ValidateResolvedConfig(ResolvedConfig{JavaCmd: "D:/missing/java.exe"})
    require.ErrorContains(t, err, "java")
}
```

- [ ] **Step 3: 运行测试确认失败**

Run: `go test ./internal/config ./internal/util -run "TestResolve|TestValidateResolvedConfig" -v`
Expected: FAIL

- [ ] **Step 4: 实现解析器与校验器**

实现逻辑：
- CLI > 项目 `.jmvn.toml` > 全局 `config.toml` > 环境变量 > PATH
- 为每个字段记录来源
- 将 `settings`、`local_repo` 解析为绝对路径
- 校验 Java 可执行文件、Maven 目录、`m2.conf`、`plexus-classworlds` JAR、可选 settings 文件

- [ ] **Step 5: 运行测试确认通过**

Run: `go test ./internal/config ./internal/util -v`
Expected: PASS

- [ ] **Step 6: 提交这一小步**

```bash
git add internal/config/resolve.go internal/config/resolve_test.go internal/util/validate.go internal/util/validate_test.go
git commit -m "feat: resolve and validate jmvn runtime config"
```

### Task 6: 构建并执行 Maven Java 启动命令

**Files:**
- Create: `internal/runner/builder.go`
- Create: `internal/runner/executor.go`
- Create: `internal/runner/builder_test.go`
- Create: `internal/runner/executor_test.go`

- [ ] **Step 1: 写命令构建测试**

```go
func TestBuildCommand_IncludesMavenLauncherAndOverrides(t *testing.T) {
    cmd := BuildCommand(cfg, []string{"clean", "install"})

    require.Equal(t, cfg.JavaCmd, cmd.Path)
    require.Contains(t, cmd.Args, "org.codehaus.plexus.classworlds.launcher.Launcher")
    require.Contains(t, cmd.Args, "--settings")
    require.Contains(t, strings.Join(cmd.Args, " "), "-Dmaven.repo.local=")
}
```

- [ ] **Step 2: 写 Maven 4 特殊参数测试**

```go
func TestBuildCommand_AddsNativeAccessForMaven4(t *testing.T) {
    cmd := BuildCommand(cfgWithMaven4, nil)
    require.Contains(t, cmd.Args, "--enable-native-access=ALL-UNNAMED")
}
```

- [ ] **Step 3: 运行测试确认失败**

Run: `go test ./internal/runner -v`
Expected: FAIL

- [ ] **Step 4: 实现命令构建与执行封装**

实现：
- `findClassworldsJar`
- `BuildCommand`
- `Exec(cmd *exec.Cmd) error`
- dry-run 需要的命令渲染辅助函数

执行器先聚焦正确返回退出码和标准 IO 透传；信号转发至少覆盖 `os.Interrupt`，平台相关差异可在实现中做条件处理。

- [ ] **Step 5: 运行 runner 测试**

Run: `go test ./internal/runner -v`
Expected: PASS

- [ ] **Step 6: 提交这一小步**

```bash
git add internal/runner/builder.go internal/runner/executor.go internal/runner/builder_test.go internal/runner/executor_test.go
git commit -m "feat: build and execute maven java launcher command"
```

### Task 7: 串起主执行链路

**Files:**
- Modify: `cmd/root.go`
- Modify: `main.go`
- Create: `cmd/root_execute_test.go`

- [ ] **Step 1: 写主链路 dry-run 测试**

```go
func TestRootCommand_DryRunPrintsResolvedJavaCommand(t *testing.T) {
    output, err := runRootForTest(t, []string{"--dry-run", "clean", "test"}, testDeps)

    require.NoError(t, err)
    require.Contains(t, output, "org.codehaus.plexus.classworlds.launcher.Launcher")
    require.Contains(t, output, "clean test")
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `go test ./cmd -run TestRootCommand_DryRunPrintsResolvedJavaCommand -v`
Expected: FAIL

- [ ] **Step 3: 实现加载配置、自动检测、解析、校验、构建命令、dry-run 与真实执行**

根命令闭环应满足：
- 从当前工作目录向上发现项目根目录
- 在项目未显式配置 JDK 时尝试自动检测
- 通过依赖注入让测试可替换文件系统/环境/执行器
- `--verbose` 打印解析过程，`--dry-run` 只输出命令不执行

- [ ] **Step 4: 运行命令层测试**

Run: `go test ./cmd -v`
Expected: PASS

- [ ] **Step 5: 运行全量单测**

Run: `go test ./...`
Expected: PASS

- [ ] **Step 6: 提交这一小步**

```bash
git add main.go cmd/root.go cmd/root_execute_test.go
git commit -m "feat: wire jmvn execution flow"
```

## Chunk 3: 辅助命令、初始化、文档与发布

### Task 8: 实现 `info`、`list`、`version` 子命令

**Files:**
- Create: `cmd/info.go`
- Create: `cmd/list.go`
- Create: `cmd/version.go`
- Create: `cmd/info_test.go`
- Create: `cmd/list_test.go`
- Create: `cmd/version_test.go`

- [ ] **Step 1: 写 `jmvn info` 输出测试**

```go
func TestInfoCommand_PrintsResolvedConfigSources(t *testing.T) {
    output := runInfoForTest(t, fakeResolvedConfig())
    require.Contains(t, output, "JDK")
    require.Contains(t, output, "[.jmvn.toml]")
}
```

- [ ] **Step 2: 写 `jmvn list` 与 `jmvn version` 测试**

```go
func TestListCommand_PrintsRegisteredToolchains(t *testing.T) {}
func TestVersionCommand_PrintsBuildVersion(t *testing.T) {}
```

- [ ] **Step 3: 运行测试确认失败**

Run: `go test ./cmd -run "TestInfoCommand|TestListCommand|TestVersionCommand" -v`
Expected: FAIL

- [ ] **Step 4: 实现三个子命令**

要求：
- `info` 展示解析结果、来源、配置文件是否存在
- `list` 展示已注册 JDK/Maven 与路径有效性
- `version` 输出版本、Go 版本、目标平台

- [ ] **Step 5: 运行命令层测试**

Run: `go test ./cmd -v`
Expected: PASS

- [ ] **Step 6: 提交这一小步**

```bash
git add cmd/info.go cmd/list.go cmd/version.go cmd/info_test.go cmd/list_test.go cmd/version_test.go
git commit -m "feat: add jmvn inspection commands"
```

### Task 9: 实现 `jmvn init` 与配置模板生成

**Files:**
- Create: `cmd/init_cmd.go`
- Create: `cmd/init_cmd_test.go`
- Create: `internal/config/template.go`
- Modify: `internal/config/global.go`
- Modify: `internal/config/project.go`

- [ ] **Step 1: 写项目级初始化测试**

```go
func TestInitCommand_WritesProjectConfigTemplate(t *testing.T) {
    runInitForTest(t, []string{"init"}, promptAnswers{
        JDK: "17",
        Maven: "3.9",
        Settings: "./maven/settings.xml",
    })

    content := readFile(t, filepath.Join(projectDir, ".jmvn.toml"))
    require.Contains(t, content, `jdk = "17"`)
}
```

- [ ] **Step 2: 写全局初始化测试**

```go
func TestInitCommand_GlobalWritesConfigToml(t *testing.T) {}
```

- [ ] **Step 3: 运行测试确认失败**

Run: `go test ./cmd -run TestInitCommand -v`
Expected: FAIL

- [ ] **Step 4: 实现交互式初始化与模板渲染**

要求：
- 支持 `jmvn init` 与 `jmvn init --global`
- 通过可替换的输入输出接口测试交互
- 项目模板只写用户填写的字段
- 全局模板包含 `defaults`、`jdks`、`mavens` 示例段

- [ ] **Step 5: 运行命令层测试**

Run: `go test ./cmd -run TestInitCommand -v`
Expected: PASS

- [ ] **Step 6: 提交这一小步**

```bash
git add cmd/init_cmd.go cmd/init_cmd_test.go internal/config/template.go internal/config/global.go internal/config/project.go
git commit -m "feat: add jmvn init command"
```

### Task 10: 完善构建、README 与 CI 发布脚本

**Files:**
- Create: `Makefile`
- Create: `.github/workflows/release.yml`
- Create: `README.md`
- Modify: `main.go`

- [ ] **Step 1: 写版本注入与构建命令验证**

手工验证项：
- `go test ./...`
- `go build -ldflags "-X main.version=1.0.0" ./...`
- `jmvn version` 输出包含注入版本

- [ ] **Step 2: 实现 Makefile 与版本变量**

要求：
- `build`、`build-all`、`clean`
- `main.go` 暴露默认版本变量供 `ldflags` 覆盖

- [ ] **Step 3: 编写 README**

至少覆盖：
- 安装方式
- 全局配置示例
- `.jmvn.toml` 示例
- 常见命令
- 原理说明与限制

- [ ] **Step 4: 增加 GitHub Actions 发布工作流**

先实现最小可用版本：
- 在 tag push 时运行测试
- 构建四个平台产物
- 上传 artifacts

- [ ] **Step 5: 完整验证**

Run: `go test ./...`
Expected: PASS

Run: `go build ./...`
Expected: PASS

- [ ] **Step 6: 提交这一小步**

```bash
git add Makefile .github/workflows/release.yml README.md main.go
git commit -m "docs: add build and release documentation"
```

## Implementation Notes

- 优先保证 `jmvn --dry-run clean install` 在一个带 `.jmvn.toml` 的示例项目中可输出正确 Java 启动命令，这是 MVP 成功标志。
- `Resolve` 与 `BuildCommand` 建议设计成可注入依赖的纯函数，避免在命令层堆积难测逻辑。
- 对 Windows 路径、空格路径、`~` 展开、相对路径要单独覆盖测试。
- `runner.Exec` 中不要直接在库函数内调用 `os.Exit`，应把退出码向上返回，由命令层统一处理，避免测试困难。
- 如果实现中发现 `init` 的交互式能力过重，可先封装模板渲染和文件写入，把真实交互留给 Cobra 命令层薄封装。

## Verification Checklist

- [ ] `go test ./...`
- [ ] `go build ./...`
- [ ] `jmvn --dry-run clean install` 能打印完整 Java 启动命令
- [ ] `jmvn info` 能显示来源追踪
- [ ] `jmvn list` 能显示已注册 JDK/Maven 的存在性
- [ ] `jmvn init` 能生成 `.jmvn.toml`

Plan complete and saved to `docs/superpowers/plans/2026-03-14-jmvn-implementation-plan.md`. Ready to execute?
