package cmd

import (
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

func loadCommandContext() (loadedCommandContext, error) {
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

	env := deps.lookupEnv()

	if projectCfg.JDK == "" && env["JMVN_JDK"] == "" {
		projectCfg.JDK = deps.detectJDKVersion(cwd)
	}

	return loadedCommandContext{
		cwd:         cwd,
		globalPath:  globalPath,
		projectPath: projectPath,
		globalCfg:   globalCfg,
		projectCfg:  projectCfg,
		env:         env,
	}, nil
}

func resolveCommandConfig() (loadedCommandContext, config.ResolvedConfig, error) {
	ctx, err := loadCommandContext()
	if err != nil {
		return loadedCommandContext{}, config.ResolvedConfig{}, err
	}

	resolved, err := deps.resolve(ctx.projectCfg, ctx.globalCfg, ctx.env, ctx.cwd)
	if err != nil {
		return loadedCommandContext{}, config.ResolvedConfig{}, err
	}
	resolved.ProjectDir = ctx.cwd
	return ctx, resolved, nil
}
