package cmd

/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/fang"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/charmtone"
	"github.com/charmbracelet/x/term"
	"github.com/spelens-gud/assert"
	termutil "github.com/spelens-gud/golangci-scope/internal/term"
	"github.com/spelens-gud/golangci-scope/internal/version"
	"github.com/spelens-gud/logger"
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
	// PersistentPreRun 运行前执行
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// 创建 Node 服务专用的 viper 实例
		rootViper = viper.New()

		assert.Then(configFile != "").Do(func() {
			// 使用命令行指定的配置文件
			rootViper.SetConfigFile(configFile)
		}).Else(func() {
			// 查找主目录
			home := assert.MustCall0RE(os.UserHomeDir, "获取用户主目录失败")
			rootViper.AddConfigPath(home)
			rootViper.AddConfigPath(".")
			rootViper.AddConfigPath("./config")
			rootViper.SetConfigType("yaml")
			rootViper.SetConfigName("default")
		})

		// 读取环境变量
		rootViper.AutomaticEnv()

		// 读取配置文件
		assert.MustCall0E(rootViper.ReadInConfig, "读取配置文件失败")

		fmt.Fprintf(os.Stderr, "使用配置文件: %s\n", rootViper.ConfigFileUsed())
	},
	// RunE 运行
	RunE: func(cmd *cobra.Command, args []string) error {
		// 创建日志实例
		logConfig := logger.LoadConfigFromViper(rootViper)
		setupDebug(logConfig)
		assert.MustCall1E(logger.InitDefault, logConfig, "创建日志实例失败")
		assert.SetLogger(logger.GetDefault())
		defer logger.GetDefault().Sync()

		err := setupAppWithProgressBar(cmd)
		if err != nil {
			return err
		}
		return err
	},
	// PersistentPostRun 运行后执行
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if debugInCISyncFile != "" {
			f, err := os.Create(debugInCISyncFile)
			if err != nil {
				logger.Fatal(err.Error())
			}
			defer f.Close()

			time.Sleep(5 * time.Second)
		}
	},
}

var rootViper *viper.Viper

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
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "配置文件路径 (默认为 ./config/default.yaml)")
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

func setupDebug(cnf *logger.Config) {
	if debug {
		cnf.Level = "debug"
		cnf.Console = true
		cnf.EnableStacktrace = true
		cnf.EnableCaller = true
	}

	return
}
