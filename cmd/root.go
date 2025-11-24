package cmd

/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/fang"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/charmtone"
	"github.com/charmbracelet/x/term"
	"github.com/spelens-gud/assert"
	"github.com/spelens-gud/golangci-scope/internal/app"
	termutil "github.com/spelens-gud/golangci-scope/internal/term"
	"github.com/spelens-gud/golangci-scope/internal/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "golangci-scope",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
	RunE: func(cmd *cobra.Command, args []string) error {
		err := setupAppWithProgressBar(cmd)
		if err != nil {
			return err
		}
		return err
	},
}

var heartbit = lipgloss.NewStyle().Foreground(charmtone.Hazy).SetString(`
  ___   __   __     __   __ _   ___  __      ____   ___  __  ____  ____ 
 / __) /  \ (  )   / _\ (  ( \ / __)(  )___ / ___) / __)/  \(  _ \(  __)
( (_ \(  O )/ (_/\/    \/    /( (__  )((___)\___ \( (__(  O )) __/ ) _) 
 \___/ \__/ \____/\_/\_/\_)__) \___)(__)    (____/ \___)\__/(__)  (____)
`)

// copied from cobra:.
const defaultVersionTemplate = `{{with .DisplayName}}{{printf "%s " .}}{{end}}{{printf "version %s" .Version}}
`

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if term.IsTerminal(os.Stdout.Fd()) {
		var b bytes.Buffer
		w := colorprofile.NewWriter(os.Stdout, os.Environ())
		w.Forward = &b
		_, _ = w.WriteString(heartbit.String())
		rootCmd.SetVersionTemplate(b.String() + "\n" + defaultVersionTemplate)
	}

	if err := fang.Execute(
		context.Background(),
		rootCmd,
		fang.WithVersion(version.Version),
		fang.WithNotifySignal(os.Interrupt),
	); err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.golangci-scope.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "run programe in debug mode")
	rootCmd.PersistentFlags().StringVar(&debugInCISyncFile, "debugcisyncfile", "", "ci sync file")
	rootCmd.PersistentFlags().StringVar(&cwd, "cwd", "", "Current working directory")
	rootCmd.PersistentFlags().StringVar(&dataDir, "data-dir", "", "Custom crush data directory")
	rootCmd.Flags().BoolVar(&help, "help", false, "Help")
	assert.MustCall1E(viper.BindPFlags, rootCmd.PersistentFlags(), "viper 隐藏失败")
}
func setupAppWithProgressBar(cmd *cobra.Command) error {
	if termutil.SupportsProgressBar() {
		_, _ = fmt.Fprintf(os.Stderr, ansi.SetIndeterminateProgressBar)
		defer func() { _, _ = fmt.Fprintf(os.Stderr, ansi.ResetProgressBar) }()
	}

	return nil
}

func setupApp(cmd *cobra.Command) (*app.App, error) {
	return app.New(cmd.Context(), nil, nil)
}
