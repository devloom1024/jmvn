package cmd

import (
	"jmvn/internal/cli"
	"jmvn/internal/config"
	"path/filepath"
)

type loadedCommandContext struct {
	cwd         string
	globalPath  string
	projectPath string
	globalCfg   config.GlobalConfig
	projectCfg  config.ProjectConfig
	env         map[string]string
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
