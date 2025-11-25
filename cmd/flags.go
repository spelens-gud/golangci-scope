package cmd

import (
	"fmt"
	"net"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	debug             bool   // debug标记
	debugInCISyncFile string // ci同步文件
	cwd               string // 当前目录
	dataDir           string // 数据目录
	help              bool   // 帮助
	configFile        string // 配置文件

	center     string    // 覆盖率中心
	agentPort  AgentPort // 覆盖率代理端口
	buildFlags string    // 构建参数
	singleton  bool      // 单一模式

	goRunExecFlag  string // go run -exec flag
	goRunArguments string // go run arguments
)

// CoverMode struct 覆盖率检测模式.
var coverMode = CoverMode{
	mode: "count",
}

func addRunFlags(cmdset *pflag.FlagSet) {
	addBuildFlags(cmdset)
	cmdset.StringVar(&goRunExecFlag, "exec", "", "same as -exec flag in 'go run' command")
	cmdset.StringVar(&goRunArguments, "arguments", "", "same as 'arguments' in 'go run' command")
	// bind to viper
	viper.BindPFlags(cmdset)
}
func addBuildFlags(cmdset *pflag.FlagSet) {
	addCommonFlags(cmdset)
	// bind to viper
	viper.BindPFlags(cmdset)
}
func addCommonFlags(cmdset *pflag.FlagSet) {
	addBasicFlags(cmdset)
	cmdset.Var(&coverMode, "mode", "coverage mode: set, count, atomic")
	cmdset.Var(&agentPort, "agentport", "a fixed port such as :8100 for registered service communicate with goc server. if not provided, using a random one")
	cmdset.BoolVar(&singleton, "singleton", false, "singleton mode, not register to goc center")
	cmdset.StringVar(&buildFlags, "buildflags", "", "specify the build flags")
	// bind to viper
	viper.BindPFlags(cmdset)
}
func addBasicFlags(cmdset *pflag.FlagSet) {
	cmdset.StringVar(&center, "center", "http://127.0.0.1:7777", "cover profile host center")
	// bind to viper
	viper.BindPFlags(cmdset)
}

// AgentPort struct 执行端口检查.
type AgentPort struct {
	port string
}

// String method 返回端口字符串.
func (agent *AgentPort) String() string {
	return agent.port
}

// Set method 设置端口.
func (agent *AgentPort) Set(v string) error {
	if v == "" {
		agent.port = ""
		return nil
	}
	_, _, err := net.SplitHostPort(v)
	if err != nil {
		return err
	}
	agent.port = v
	return nil
}

// Type method 返回端口类型.
func (agent *AgentPort) Type() string {
	return "string"
}

// CoverMode struct 覆盖率检测模式.
type CoverMode struct {
	mode string
}

// String method 返回检测模式字符串.
func (m *CoverMode) String() string {
	return m.mode
}

// Set method 设置检测模式.
func (m *CoverMode) Set(v string) error {
	if v == "" {
		m.mode = "count"
		return nil
	}
	if v != "set" && v != "count" && v != "atomic" {
		return fmt.Errorf("unknown mode")
	}
	m.mode = v
	return nil
}

// Type method 获取检测模式类型.
func (m *CoverMode) Type() string {
	return "string"
}
