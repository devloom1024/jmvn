# jmvn Command Design Fixes Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 修复 `jmvn` 当前命令层的所有已知设计缺陷，使执行路径、信息展示、初始化、help 文案和子命令参数模型保持一致。

**Architecture:** 先把“命令上下文加载与解析”抽成单一共享流程，再让 `root`、`info`、`list`、`init` 都围绕这个流程工作，避免再出现执行路径和展示路径分叉。随后把根命令覆盖参数升级为持久参数，并补齐 `init --global` 的真实配置模型与 help/README 示例，使用户能从 help 直接发现和正确使用这些能力。

**Tech Stack:** Go, Cobra,现有 `cmd`/`internal/config` 包，`go test`

---

## File Structure

**Modify:**
- `cmd/root.go`: 把根命令 flags 改为持久参数，抽出共享解析上下文/解析结果装载函数，消除 `root` 与 `info` 的重复流程。
- `cmd/info.go`: 改为复用共享解析流程，确保输出与实际执行一致。
- `cmd/list.go`: 补充更清晰的 help/long/example 文案；必要时复用共享全局配置读取逻辑。
- `cmd/init_cmd.go`: 重做 `init --global` 交互模型，避免产出不可执行或误导性的配置。
- `cmd/version.go`: 仅在需要统一 help 模板/示例时调整命令元数据，不改业务行为。
- `internal/config/template.go`: 支持渲染完整且自洽的全局初始化模板。
- `README.md`: 同步新的 help 行为、`info` 语义和 `init --global` 示例。

**Create:**
- `cmd/runtime.go`: 放共享的命令上下文装载/解析函数，避免继续在多个子命令里复制 `cwd -> load config -> detect -> resolve -> validate`。
- `cmd/help_test.go`: 覆盖根帮助输出、`init --help` 输出以及根持久参数在子命令帮助中的可见性。

**Modify Tests:**
- `cmd/root_test.go`
- `cmd/root_detect_test.go`
- `cmd/root_verbose_test.go`
- `cmd/info_test.go`
- `cmd/init_cmd_test.go`
- `cmd/list_test.go`

---

## Chunk 1: Unified Command Resolution

### Task 1: 定义共享命令上下文结构

**Files:**
- Create: `cmd/runtime.go`
- Test: `cmd/info_test.go`

- [ ] **Step 1: 写失败测试，说明 `info` 必须和执行路径使用同一套解析输入**

```go
func TestInfoCommand_UsesDetectedJDKWhenProjectConfigMissing(t *testing.T) {
	original := deps
	defer func() { deps = original }()

	projectDir := t.TempDir()
	capturedProjectJDK := ""
	deps = commandDeps{
		getwd:            func() (string, error) { return projectDir, nil },
		userHomeDir:      func() string { return `D:/home` },
		loadGlobal:       func(string) (config.GlobalConfig, error) { return config.GlobalConfig{}, nil },
		loadProject:      func(string) (config.ProjectConfig, error) { return config.ProjectConfig{}, nil },
		detectJDKVersion: func(string) string { return "8" },
		resolve: func(cliOpts cli.Options, projectCfg config.ProjectConfig, globalCfg config.GlobalConfig, env map[string]string, projectDir string) (config.ResolvedConfig, error) {
			capturedProjectJDK = projectCfg.JDK
			return config.ResolvedConfig{JavaCmd: `java`, MavenHome: `maven`, ProjectDir: projectDir}, nil
		},
		lookupEnv:  func() map[string]string { return map[string]string{} },
		promptInit: func(bool) (promptAnswers, error) { return promptAnswers{}, nil },
	}

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"info"})
	_ = cmd.Execute()

	if capturedProjectJDK != "8" {
		t.Fatalf("expected info to reuse detected JDK, got %q", capturedProjectJDK)
	}
}
```

- [ ] **Step 2: 跑测试确认当前失败**

Run: `go test ./cmd -run TestInfoCommand_UsesDetectedJDKWhenProjectConfigMissing -v`
Expected: FAIL，`capturedProjectJDK` 为空字符串。

- [ ] **Step 3: 在 `cmd/runtime.go` 引入共享装载函数**

```go
type loadedCommandContext struct {
	cwd        string
	globalPath string
	projectPath string
	globalCfg  config.GlobalConfig
	projectCfg config.ProjectConfig
	env        map[string]string
}

func loadCommandContext(state *executionState) (loadedCommandContext, error) {
	cwd, err := deps.getwd()
	if err != nil {
		return loadedCommandContext{}, err
	}

	globalPath := filepath.Join(deps.userHomeDir(), ".jmvn", "config.toml")
	projectPath := filepath.Join(cwd, ".jmvn.toml")
	globalCfg, err := deps.loadGlobal(globalPath)
	if err != nil {
		return loadedCommandContext{}, err
	}
	projectCfg, err := deps.loadProject(projectPath)
	if err != nil {
		return loadedCommandContext{}, err
	}

	opts := cli.Options{}
	if state != nil {
		opts = state.options
	}
	if projectCfg.JDK == "" && opts.JDK == "" {
		projectCfg.JDK = deps.detectJDKVersion(cwd)
	}

	return loadedCommandContext{
		cwd:         cwd,
		globalPath:  globalPath,
		projectPath: projectPath,
		globalCfg:   globalCfg,
		projectCfg:  projectCfg,
		env:         deps.lookupEnv(),
	}, nil
}
```

- [ ] **Step 4: 在共享函数基础上补一个统一解析函数**

```go
func resolveCommandConfig(state *executionState) (loadedCommandContext, config.ResolvedConfig, error) {
	ctx, err := loadCommandContext(state)
	if err != nil {
		return loadedCommandContext{}, config.ResolvedConfig{}, err
	}

	opts := cli.Options{}
	if state != nil {
		opts = state.options
	}
	resolved, err := deps.resolve(opts, ctx.projectCfg, ctx.globalCfg, ctx.env, ctx.cwd)
	if err != nil {
		return loadedCommandContext{}, config.ResolvedConfig{}, err
	}
	resolved.ProjectDir = ctx.cwd
	return ctx, resolved, nil
}
```

- [ ] **Step 5: 跑测试确认通过**

Run: `go test ./cmd -run TestInfoCommand_UsesDetectedJDKWhenProjectConfigMissing -v`
Expected: PASS

- [ ] **Step 6: 提交**

```bash
git add cmd/runtime.go cmd/info_test.go
git commit -m "refactor: share command resolution flow"
```

### Task 2: 让 root 和 info 共用统一解析流程

**Files:**
- Modify: `cmd/root.go`
- Modify: `cmd/info.go`
- Test: `cmd/root_detect_test.go`
- Test: `cmd/info_test.go`

- [ ] **Step 1: 写失败测试，要求 `info` 输出与 root 解析来源一致**

```go
func TestInfoCommand_PrintsDetectedSourceWhenProjectConfigMissing(t *testing.T) {
	original := deps
	defer func() { deps = original }()

	projectDir := t.TempDir()
	deps = commandDeps{
		getwd:            func() (string, error) { return projectDir, nil },
		userHomeDir:      func() string { return `D:/home` },
		loadGlobal:       func(string) (config.GlobalConfig, error) { return config.GlobalConfig{}, nil },
		loadProject:      func(string) (config.ProjectConfig, error) { return config.ProjectConfig{}, nil },
		detectJDKVersion: func(string) string { return "8" },
		resolve: func(cliOpts cli.Options, projectCfg config.ProjectConfig, globalCfg config.GlobalConfig, env map[string]string, projectDir string) (config.ResolvedConfig, error) {
			return config.ResolvedConfig{
				JavaCmd:       filepath.Clean(`D:/jdks/jdk-8/bin/java`),
				MavenHome:     filepath.Clean(`D:/mavens/apache-maven-3.9.6`),
				JavaCmdSource: "project",
			}, nil
		},
		lookupEnv:  func() map[string]string { return map[string]string{} },
		promptInit: func(bool) (promptAnswers, error) { return promptAnswers{}, nil },
	}

	cmd := NewRootCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stdout)
	cmd.SetArgs([]string{"info"})

	_ = cmd.Execute()
	if !strings.Contains(stdout.String(), "[project]") {
		t.Fatalf("expected detected project source in info output, got %q", stdout.String())
	}
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./cmd -run "TestInfoCommand_PrintsDetectedSourceWhenProjectConfigMissing|TestRootCommand_UsesDetectedJDKWhenProjectConfigMissing" -v`
Expected: `info` 相关测试 FAIL，root 相关测试保持 PASS。

- [ ] **Step 3: 将 `runRootCommand` 改为调用 `resolveCommandConfig`**

```go
func runRootCommand(cmd *cobra.Command, state *executionState) error {
	_, resolved, err := resolveCommandConfig(state)
	if err != nil {
		return err
	}
	if err := deps.validateResolved(resolved); err != nil {
		return err
	}
	command, err := deps.buildCommand(resolved, state.mavenArgs)
	if err != nil {
		return err
	}
	// 保留 verbose/dry-run/execute 分支
}
```

- [ ] **Step 4: 将 `info` 改为调用同一个共享解析函数**

```go
func newInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Show the effective jmvn resolution",
		RunE: func(cmd *cobra.Command, args []string) error {
			state := executionStateFromContext(cmd.Root().Context())
			ctx, resolved, err := resolveCommandConfig(state)
			if err != nil {
				return err
			}
			return printInfo(cmd, ctx, resolved)
		},
	}
}
```

- [ ] **Step 5: 跑测试确认通过**

Run: `go test ./cmd -run "TestInfoCommand_.*|TestRootCommand_UsesDetectedJDKWhenProjectConfigMissing" -v`
Expected: PASS

- [ ] **Step 6: 提交**

```bash
git add cmd/root.go cmd/info.go cmd/root_detect_test.go cmd/info_test.go
git commit -m "fix: align info command with execution resolution"
```

---

## Chunk 2: Persistent Flags And Help Model

### Task 3: 把覆盖参数升级为根持久参数

**Files:**
- Modify: `cmd/root.go`
- Test: `cmd/root_test.go`
- Test: `cmd/help_test.go`

- [ ] **Step 1: 写失败测试，要求子命令继承 `--jdk`**

```go
func TestInfoCommand_AcceptsPersistentRootFlags(t *testing.T) {
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"info", "--jdk", "8"})

	_, _, err := executeForTest(cmd)
	if err != nil {
		t.Fatalf("expected inherited root flag support, got %v", err)
	}
}
```

- [ ] **Step 2: 跑测试确认当前失败**

Run: `go test ./cmd -run TestInfoCommand_AcceptsPersistentRootFlags -v`
Expected: FAIL，报 `unknown flag: --jdk`。

- [ ] **Step 3: 将根参数迁移到 `PersistentFlags()`**

```go
flags := cmd.PersistentFlags()
flags.StringVarP(&state.options.JDK, "jdk", "j", "", "Override JDK version")
flags.StringVarP(&state.options.Maven, "maven", "m", "", "Override Maven version")
flags.StringVarP(&state.options.Settings, "settings", "s", "", "Override settings.xml path")
flags.StringVarP(&state.options.LocalRepo, "local-repo", "r", "", "Override local Maven repository path")
flags.BoolVarP(&state.options.DryRun, "dry-run", "n", false, "Print the resolved Java command without executing it")
flags.BoolVarP(&state.options.Verbose, "verbose", "v", false, "Print verbose resolution output")
```

- [ ] **Step 4: 跑测试确认通过**

Run: `go test ./cmd -run "TestInfoCommand_AcceptsPersistentRootFlags|TestRootCommand_ParsesOwnFlagsAndLeavesMavenArgs" -v`
Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add cmd/root.go cmd/root_test.go cmd/help_test.go
git commit -m "fix: make jmvn overrides persistent flags"
```

### Task 4: 改造 help 文案，让 `init --global` 和常见用法在根帮助可发现

**Files:**
- Modify: `cmd/root.go`
- Modify: `cmd/init_cmd.go`
- Create: `cmd/help_test.go`
- Modify: `README.md`

- [ ] **Step 1: 写失败测试，要求根帮助展示 `init --global` 示例**

```go
func TestRootHelp_IncludesInitGlobalExample(t *testing.T) {
	cmd := NewRootCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stdout)
	cmd.SetArgs([]string{"--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(stdout.String(), "jmvn init --global") {
		t.Fatalf("expected init --global example in root help, got %q", stdout.String())
	}
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./cmd -run TestRootHelp_IncludesInitGlobalExample -v`
Expected: FAIL

- [ ] **Step 3: 为根命令和 `init` 子命令补齐 `Long`/`Example`**

```go
cmd := &cobra.Command{
	Use:   "jmvn [maven-args...]",
	Short: "Resolve JDK and Maven, then launch Maven with the selected Java runtime",
	Long: `jmvn merges CLI flags, project config, global config and environment,
then resolves the effective JDK, Maven, settings.xml and local repository.`,
	Example: strings.Join([]string{
		"jmvn --dry-run clean test",
		"jmvn info --jdk 8",
		"jmvn init --global",
	}, "\n"),
}
```

```go
cmd := &cobra.Command{
	Use:   "init",
	Short: "Initialize jmvn configuration",
	Long:  "Create a project-local .jmvn.toml or a global ~/.jmvn/config.toml file.",
	Example: strings.Join([]string{
		"jmvn init",
		"jmvn init --global",
	}, "\n"),
}
```

- [ ] **Step 4: 跑测试确认通过**

Run: `go test ./cmd -run "TestRootHelp_IncludesInitGlobalExample|TestVersionCommand_PrintsBuildVersion" -v`
Expected: PASS

- [ ] **Step 5: 更新 README 示例**

```md
常见命令：

jmvn --dry-run clean test
jmvn info --jdk 8
jmvn init --global
jmvn init --help
```

- [ ] **Step 6: 提交**

```bash
git add cmd/root.go cmd/init_cmd.go cmd/help_test.go README.md
git commit -m "docs: improve jmvn help and examples"
```

---

## Chunk 3: Fix `init --global` Configuration Model

### Task 5: 让 `init --global` 生成可执行且自洽的全局配置

**Files:**
- Modify: `cmd/init_cmd.go`
- Modify: `internal/config/template.go`
- Modify: `cmd/init_cmd_test.go`

- [ ] **Step 1: 写失败测试，要求全局初始化写出默认 JDK 对应的注册映射**

```go
func TestInitCommand_GlobalWritesRegisteredDefaultJDK(t *testing.T) {
	original := deps
	defer func() { deps = original }()

	homeDir := t.TempDir()
	deps = commandDeps{
		userHomeDir: func() string { return homeDir },
		promptInit: func(global bool) (promptAnswers, error) {
			return promptAnswers{
				JDK:       "17",
				JDKHome:   `D:/jdks/jdk-17`,
				Maven:     "3.9",
				MavenHome: `D:/mavens/apache-maven-3.9.6`,
				Settings:  `D:/users/demo/.m2/settings.xml`,
				LocalRepo: `D:/users/demo/.m2/repository`,
			}, nil
		},
	}

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"init", "--global"})
	_ = cmd.Execute()

	content, _ := os.ReadFile(filepath.Join(homeDir, ".jmvn", "config.toml"))
	text := string(content)
	if !strings.Contains(text, `"17" = "D:/jdks/jdk-17"`) {
		t.Fatalf("expected registered default JDK mapping, got %q", text)
	}
}
```

- [ ] **Step 2: 跑测试确认当前失败**

Run: `go test ./cmd -run TestInitCommand_GlobalWritesRegisteredDefaultJDK -v`
Expected: FAIL

- [ ] **Step 3: 扩展 `promptAnswers`，支持全局初始化的注册信息**

```go
type promptAnswers struct {
	JDK       string
	JDKHome   string
	Maven     string
	MavenHome string
	Settings  string
	LocalRepo string
}
```

- [ ] **Step 4: 更新 `defaultPromptInit(true)`，采集默认版本和对应安装目录**

```go
if answers.JDK, err = ask("Default JDK version: "); err != nil { ... }
if answers.JDKHome, err = ask("JDK home for that version: "); err != nil { ... }
if answers.Maven, err = ask("Default Maven version (optional): "); err != nil { ... }
if answers.MavenHome, err = ask("Default Maven home: "); err != nil { ... }
```

- [ ] **Step 5: 让 `RenderGlobalConfig` 产出自洽模板**

```go
func RenderGlobalConfig(jdk string, jdkHome string, maven string, mavenHome string, settings string, localRepo string) string {
	lines := []string{"[defaults]"}
	if jdk != "" {
		lines = append(lines, fmt.Sprintf("jdk = %q", jdk))
	}
	if mavenHome != "" {
		lines = append(lines, fmt.Sprintf("maven_home = %q", mavenHome))
	}
	// ...
	lines = append(lines, "", "[jdks]")
	if jdk != "" && jdkHome != "" {
		lines = append(lines, fmt.Sprintf("%q = %q", jdk, jdkHome))
	}
	lines = append(lines, "", "[mavens]")
	if maven != "" && mavenHome != "" {
		lines = append(lines, fmt.Sprintf("%q = %q", maven, mavenHome))
	}
	return strings.Join(lines, "\n") + "\n"
}
```

- [ ] **Step 6: 跑测试确认通过**

Run: `go test ./cmd -run "TestInitCommand_Global.*|TestInitCommand_WritesProjectConfigTemplate" -v`
Expected: PASS

- [ ] **Step 7: 提交**

```bash
git add cmd/init_cmd.go internal/config/template.go cmd/init_cmd_test.go
git commit -m "fix: generate self-consistent global init config"
```

### Task 6: 给 `init` 增加覆写保护和更明确的模式提示

**Files:**
- Modify: `cmd/init_cmd.go`
- Modify: `cmd/init_cmd_test.go`
- Modify: `cmd/help_test.go`

- [ ] **Step 1: 写失败测试，要求已有配置文件时给出明确错误**

```go
func TestInitCommand_ProjectRefusesToOverwriteExistingConfig(t *testing.T) {
	projectDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(projectDir, ".jmvn.toml"), []byte("jdk = \"17\"\n"), 0o644)

	original := deps
	defer func() { deps = original }()
	deps = commandDeps{
		getwd: func() (string, error) { return projectDir, nil },
		promptInit: func(bool) (promptAnswers, error) { return promptAnswers{}, nil },
	}

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"init"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("expected overwrite protection, got %v", err)
	}
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./cmd -run TestInitCommand_ProjectRefusesToOverwriteExistingConfig -v`
Expected: FAIL

- [ ] **Step 3: 在 `writeProjectConfig` / `writeGlobalConfig` 增加存在性检查**

```go
if _, err := os.Stat(path); err == nil {
	return fmt.Errorf("configuration already exists: %s", path)
}
```

- [ ] **Step 4: 在 `init --help` 的 `Long`/`Example` 中说明 project/global 两种模式**

```go
Long: "Create a project-local .jmvn.toml or a global ~/.jmvn/config.toml file. Use --global to initialize the shared toolchain registry.",
```

- [ ] **Step 5: 跑测试确认通过**

Run: `go test ./cmd -run "TestInitCommand_ProjectRefusesToOverwriteExistingConfig|TestRootHelp_IncludesInitGlobalExample" -v`
Expected: PASS

- [ ] **Step 6: 提交**

```bash
git add cmd/init_cmd.go cmd/init_cmd_test.go cmd/help_test.go
git commit -m "fix: harden init command behavior"
```

---

## Chunk 4: Final Validation And Documentation

### Task 7: 覆盖命令设计回归测试并完成文档同步

**Files:**
- Modify: `cmd/info_test.go`
- Modify: `cmd/root_test.go`
- Modify: `cmd/init_cmd_test.go`
- Modify: `cmd/list_test.go`
- Modify: `README.md`

- [ ] **Step 1: 补齐最终回归测试矩阵**

```go
// 重点覆盖：
// 1. jmvn info 使用 detectJDKVersion
// 2. jmvn info --jdk 8 可以工作
// 3. jmvn --help 包含 init --global 示例
// 4. jmvn init --global 生成可执行配置
// 5. init 不覆写已有文件
```

- [ ] **Step 2: 跑 cmd 包测试**

Run: `go test ./cmd -v`
Expected: PASS

- [ ] **Step 3: 跑整仓测试**

Run: `go test ./...`
Expected: PASS

- [ ] **Step 4: 手工验证 help 和配置输出**

Run: `jmvn --help`
Expected: 输出包含 `jmvn init --global`

Run: `jmvn init --help`
Expected: 输出说明 `--global` 会初始化全局配置

Run: `jmvn info --jdk 8`
Expected: 不再报 `unknown flag: --jdk`

- [ ] **Step 5: 提交**

```bash
git add cmd README.md
git commit -m "test: cover jmvn command design regressions"
```

