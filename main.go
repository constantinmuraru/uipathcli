// Package main contains the entry point of the uipath command line interface.
//
// It only initializes the different packages and delegates the actual work.
package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/UiPath/uipathcli/auth"
	"github.com/UiPath/uipathcli/cache"
	"github.com/UiPath/uipathcli/commandline"
	"github.com/UiPath/uipathcli/config"
	"github.com/UiPath/uipathcli/executor"
	"github.com/UiPath/uipathcli/parser"
	"github.com/UiPath/uipathcli/plugin"
	plugin_digitizer "github.com/UiPath/uipathcli/plugin/digitizer"
	plugin_orchestrator "github.com/UiPath/uipathcli/plugin/orchestrator"
	"github.com/UiPath/uipathcli/utils"
)

func authenticators(pluginsCfg config.PluginConfig) []auth.Authenticator {
	authenticators := []auth.Authenticator{}
	for _, authenticator := range pluginsCfg.Authenticators {
		authenticators = append(authenticators, auth.NewExternalAuthenticator(
			*auth.NewExternalAuthenticatorConfig(authenticator.Name, authenticator.Path),
		))
	}
	return append(authenticators,
		auth.NewPatAuthenticator(),
		auth.NewOAuthAuthenticator(cache.NewFileCache(), auth.NewExecBrowserLauncher()),
		auth.NewBearerAuthenticator(cache.NewFileCache()),
	)
}

func colorsSupported() bool {
	_, noColorSet := os.LookupEnv("NO_COLOR")
	term, _ := os.LookupEnv("TERM")
	omitColors := noColorSet || term == "dumb" || runtime.GOOS == "windows"
	return !omitColors
}

func stdIn() utils.Stream {
	return utils.NewReaderStream(parser.RawBodyParameterName, os.Stdin)
}

func main() {
	configProvider := config.NewConfigProvider(
		config.NewConfigFileStore(os.Getenv("UIPATH_CONFIGURATION_PATH")),
	)
	err := configProvider.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(131)
	}
	pluginConfigProvider := config.NewPluginConfigProvider(
		config.NewPluginConfigFileStore(os.Getenv("UIPATH_PLUGINS_PATH")),
	)
	err = pluginConfigProvider.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(132)
	}

	pluginConfig := pluginConfigProvider.Config()
	authenticators := authenticators(pluginConfig)
	cli := commandline.NewCli(
		os.Stdin,
		os.Stdout,
		os.Stderr,
		colorsSupported(),
		*commandline.NewDefinitionProvider(
			commandline.NewDefinitionFileStore(os.Getenv("UIPATH_DEFINITIONS_PATH")),
			parser.NewOpenApiParser(),
			[]plugin.CommandPlugin{
				plugin_digitizer.DigitizeCommand{},
				plugin_orchestrator.UploadCommand{},
				plugin_orchestrator.DownloadCommand{},
			},
		),
		*configProvider,
		executor.NewHttpExecutor(authenticators),
		executor.NewPluginExecutor(authenticators),
	)

	input := stdIn()
	err = cli.Run(os.Args, input)
	if err != nil {
		os.Exit(1)
	}
}
