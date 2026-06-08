package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"jmvn/internal/cli"
	"jmvn/internal/config"
	"jmvn/internal/detect"
	"jmvn/internal/runner"
	"jmvn/internal/validate"

	"github.com/spf13/cobra"
)

type executionState struct {
	mavenArgs []string
	dryRun    bool
}

type promptAnswers struct {
	JDK       string
	JDKHome   string
	Maven     string
	Settings  string
	LocalRepo string
	MavenHome string
}

type commandDeps struct {
	getwd            func() (string, error)
	userHomeDir      func() string
	loadGlobal       func(string) (config.GlobalConfig, error)
	loadProject      func(string) (config.ProjectConfig, error)
	detectJDKVersion func(string) string
	resolve          func(config.ProjectConfig, config.GlobalConfig, map[string]string, string) (config.ResolvedConfig, error)
	validateResolved func(config.ResolvedConfig) error
	buildCommand     func(config.ResolvedConfig, []string) (*exec.Cmd, error)
	lookupEnv        func() map[string]string
	promptInit       func(bool) (promptAnswers, error)
	executeCommand   func(*exec.Cmd) error
}

var deps = commandDeps{
	getwd:       os.Getwd,
	userHomeDir: userHomeDir,
	loadGlobal: func(path string) (config.GlobalConfig, error) {
		cfg, err := config.LoadGlobal(path)
		if err != nil && config.IsNotExist(err) {
			return config.GlobalConfig{}, nil
		}
		return cfg, err
	},
	loadProject: func(path string) (config.ProjectConfig, error) {
		cfg, err := config.LoadProject(path)
		if err != nil && config.IsNotExist(err) {
			return config.ProjectConfig{}, nil
		}
		return cfg, err
	},
	detectJDKVersion: detect.DetectJDKVersion,
	resolve:          config.Resolve,
	validateResolved: validate.ResolvedConfig,
	buildCommand:     runner.BuildCommand,
	lookupEnv: func() map[string]string {
		result := map[string]string{}
		for _, key := range []string{
			"JAVA_HOME", "MAVEN_HOME", "M2_HOME",
			"JMVN_JDK", "JMVN_MAVEN", "JMVN_MAVEN_HOME", "JMVN_SETTINGS", "JMVN_LOCAL_REPO",
		} {
			if value := os.Getenv(key); value != "" {
				result[key] = value
			}
		}
		return result
	},
	promptInit: defaultPromptInit,
	executeCommand: func(cmd *exec.Cmd) error {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return runner.Exec(cmd)
	},
}

func NewRootCmd() *cobra.Command {
	state := &executionState{}
	cmd := &cobra.Command{
		Use:   "jmvn [maven-args...]",
		Short: "Run Maven with the resolved JDK",
		Long: `jmvn merges project config, global config and environment,
then resolves the effective JDK, Maven, settings.xml and local repository.

All arguments are passed directly to Maven unless the first argument
starts with ":" — those are jmvn's own commands:

  :init        Initialize project or global configuration
  :info        Show resolved JDK / Maven / settings
  :list        List registered JDK and Maven toolchains
  :version     Print jmvn version
  :dry-run     Show the resolved Java command without executing it
  :help        Show this help

Examples:
  jmvn clean install
  jmvn -pl module -am test
  jmvn :info
  jmvn :dry-run clean test
  jmvn :init`,
		Example: strings.Join([]string{
			"  jmvn clean install",
			"  jmvn -pl flight-ticket-bdos-business -am test",
			"  JMVN_JDK=17 jmvn clean test",
			"  jmvn :info",
			"  jmvn :dry-run clean test",
			"  jmvn :init",
			"  jmvn :init --global",
		}, "\n"),
		DisableFlagParsing: true,
		SilenceUsage:       true,
		SilenceErrors:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 && strings.HasPrefix(args[0], ":") {
				return handleJmvnCommand(cmd, state, args)
			}
			state.mavenArgs = append([]string(nil), args...)
			state.dryRun = false
			return runMaven(cmd, state)
		},
	}

	cmd.SetContext(withExecutionState(cmd.Context(), state))
	return cmd
}

func handleJmvnCommand(cmd *cobra.Command, state *executionState, args []string) error {
	command := args[0]
	remaining := args[1:]

	switch command {
	case ":init":
		return handleInit(cmd, remaining)
	case ":info":
		return handleInfo(cmd)
	case ":list":
		return handleList(cmd)
	case ":version":
		return handleVersion(cmd)
	case ":dry-run":
		state.mavenArgs = append([]string(nil), remaining...)
		state.dryRun = true
		return runMaven(cmd, state)
	case ":help":
		return cmd.Help()
	default:
		return fmt.Errorf("未知的 jmvn 命令: %s（可用: :init, :info, :list, :version, :dry-run, :help）", command)
	}
}

func runMaven(cmd *cobra.Command, state *executionState) error {
	stripLeadingMvnPrefix(&state.mavenArgs)

	_, resolved, err := resolveCommandConfig()
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
	if state.dryRun {
		_, err = fmt.Fprintln(cmd.OutOrStdout(), renderCommand(command))
		return err
	}
	return deps.executeCommand(command)
}

func executeForTest(cmd *cobra.Command) (cli.Options, []string, error) {
	original := deps
	deps = commandDeps{
		getwd:            func() (string, error) { return `D:/test`, nil },
		userHomeDir:      func() string { return `D:/home` },
		loadGlobal:       func(string) (config.GlobalConfig, error) { return config.GlobalConfig{}, nil },
		loadProject:      func(string) (config.ProjectConfig, error) { return config.ProjectConfig{}, nil },
		detectJDKVersion: detect.DetectJDKVersion,
		resolve: func(config.ProjectConfig, config.GlobalConfig, map[string]string, string) (config.ResolvedConfig, error) {
			return config.ResolvedConfig{JavaCmd: `java`, MavenHome: `maven`, ProjectDir: `D:/test`}, nil
		},
		validateResolved: func(config.ResolvedConfig) error { return nil },
		buildCommand: func(cfg config.ResolvedConfig, mavenArgs []string) (*exec.Cmd, error) {
			command := exec.Command(cfg.JavaCmd, mavenArgs...)
			command.Dir = cfg.ProjectDir
			return command, nil
		},
		lookupEnv:      func() map[string]string { return map[string]string{} },
		promptInit:     func(bool) (promptAnswers, error) { return promptAnswers{}, nil },
		executeCommand: func(*exec.Cmd) error { return nil },
	}
	defer func() { deps = original }()

	err := cmd.Execute()
	state := executionStateFromContext(cmd.Context())
	if state == nil {
		return cli.Options{}, nil, err
	}
	return cli.Options{}, append([]string(nil), state.mavenArgs...), err
}

func userHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home
}

func renderCommand(cmd *exec.Cmd) string {
	return strings.Join(append([]string{cmd.Path}, cmd.Args[1:]...), " ")
}

func bracketSource(source string) string {
	if source == "" {
		return ""
	}
	return source
}

func stripLeadingMvnPrefix(args *[]string) {
	if len(*args) > 0 && ((*args)[0] == "mvn" || (*args)[0] == "mvnw") {
		*args = (*args)[1:]
	}
}
