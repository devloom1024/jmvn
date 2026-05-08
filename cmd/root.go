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
	options   cli.Options
	mavenArgs []string
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
	resolve          func(cli.Options, config.ProjectConfig, config.GlobalConfig, map[string]string, string) (config.ResolvedConfig, error)
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
		for _, key := range []string{"JAVA_HOME", "MAVEN_HOME", "M2_HOME"} {
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
		Short: "Resolve JDK and Maven, then launch Maven with the selected Java runtime",
		Long: `jmvn merges CLI flags, project config, global config and environment,
then resolves the effective JDK, Maven, settings.xml and local repository.`,
		Example: strings.Join([]string{
			"  jmvn --dry-run clean test",
			"  jmvn info --jdk 8",
			"  jmvn init --global",
		}, "\n"),
		SilenceUsage:  true,
		SilenceErrors: true,
		Args: func(cmd *cobra.Command, args []string) error {
			state.mavenArgs = append([]string(nil), args...)
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			state.mavenArgs = append([]string(nil), args...)
			return runRootCommand(cmd, state)
		},
	}

	flags := cmd.PersistentFlags()
	flags.SetInterspersed(false)
	cmd.Flags().SetInterspersed(false)
	flags.StringVarP(&state.options.JDK, "jdk", "j", "", "Override JDK version")
	flags.StringVarP(&state.options.Maven, "maven", "m", "", "Override Maven version")
	flags.StringVarP(&state.options.Settings, "settings", "s", "", "Override settings.xml path")
	flags.StringVarP(&state.options.LocalRepo, "local-repo", "r", "", "Override local Maven repository path")
	flags.BoolVarP(&state.options.DryRun, "dry-run", "n", false, "Print the resolved Java command without executing it")
	flags.BoolVarP(&state.options.Verbose, "verbose", "v", false, "Print verbose resolution output")

	cmd.AddCommand(newVersionCmd())
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newInfoCmd())
	cmd.AddCommand(newInitCmd())
	cmd.SetContext(withExecutionState(cmd.Context(), state))
	return cmd
}

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
	if state.options.Verbose {
		if err := printVerboseResolution(cmd, resolved); err != nil {
			return err
		}
	}
	if state.options.DryRun {
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
		resolve: func(cli.Options, config.ProjectConfig, config.GlobalConfig, map[string]string, string) (config.ResolvedConfig, error) {
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
	return state.options, append([]string(nil), state.mavenArgs...), err
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

func printVerboseResolution(cmd *cobra.Command, resolved config.ResolvedConfig) error {
	_, err := fmt.Fprintf(
		cmd.OutOrStdout(),
		"%s\n%s %s [%s]\n%s %s [%s]\n%s %s [%s]\n%s %s [%s]\n%s %s\n",
		styledHeader("jmvn resolution"),
		styledLabel("JDK       "), resolved.JavaCmd, bracketSource(resolved.JavaCmdSource),
		styledLabel("Maven     "), resolved.MavenHome, bracketSource(resolved.MavenHomeSource),
		styledLabel("Settings  "), resolved.Settings, bracketSource(resolved.SettingsSource),
		styledLabel("Local Repo"), resolved.LocalRepo, bracketSource(resolved.LocalRepoSource),
		styledLabel("Project Dir"), resolved.ProjectDir,
	)
	return err
}

func bracketSource(source string) string {
	if source == "" {
		return ""
	}
	return source
}
