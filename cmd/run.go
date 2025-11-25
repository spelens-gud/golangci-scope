package cmd

/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"

	"github.com/spelens-gud/golangci-scope/internal/build"
	"github.com/spelens-gud/golangci-scope/internal/cover"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// runCmd represents the run command.
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		wd, err := os.Getwd()
		if err != nil {
			log.Fatalf("Fail to build: %v", err)
		}
		gocBuild, err := build.NewBuild(buildFlags, args, wd, buildOutput)
		if err != nil {
			log.Fatalf("Fail to run: %v", err)
		}
		gocBuild.GoRunExecFlag = goRunExecFlag
		gocBuild.GoRunArguments = goRunArguments
		defer gocBuild.Clean()

		server := cover.NewMemoryBasedServer() // only save services in memory

		// start goc server
		var l = newLocalListener(agentPort.String())
		go func() {
			err = server.Route(ioutil.Discard).RunListener(l)
			if err != nil {
				log.Fatalf("Start goc server failed: %v", err)
			}
		}()
		gocServer := fmt.Sprintf("http://%s", l.Addr().String())
		fmt.Printf("[goc] goc server started: %s \n", gocServer)

		if viper.IsSet("center") {
			gocServer = center
		}

		// execute covers for the target source with original buildFlags and new GOPATH( tmp:original )
		ci := &cover.CoverInfo{
			Args:                     buildFlags,
			GoPath:                   gocBuild.NewGOPATH,
			Target:                   gocBuild.TmpDir,
			Mode:                     coverMode.String(),
			Center:                   gocServer,
			Singleton:                singleton,
			AgentPort:                "",
			IsMod:                    gocBuild.IsMod,
			ModRootPath:              gocBuild.ModRootPath,
			OneMainPackage:           true, // go run is similar with go build, build only one main package
			GlobalCoverVarImportPath: gocBuild.GlobalCoverVarImportPath,
		}
		err = cover.Execute(ci)
		if err != nil {
			log.Fatalf("Fail to run: %v", err)
		}

		if err := gocBuild.Run(); err != nil {
			log.Fatalf("Fail to run: %v", err)
		}
	},
}

func init() {
	addRunFlags(runCmd.Flags())
	rootCmd.AddCommand(runCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
func newLocalListener(addr string) net.Listener {
	if addr == "" {
		addr = "127.0.0.1:0"
	}
	l, err := net.Listen("tcp", addr)
	if err != nil {
		if l, err = net.Listen("tcp6", "[::1]:0"); err != nil {
			log.Fatalf("failed to listen on a port: %v", err)
		}
	}
	return l
}
