package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/cuipeiyu/go-i18n"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

const configFileName = ".goi18n.yaml"
const configFileType = "yaml"

type M map[string]i18n.Message

func (m M) write2File(path, source string) error {
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		_ = os.MkdirAll(dir, os.ModeDir|os.ModePerm)
	}

	t := make(map[string]map[string]string)
	for k, v := range m {
		// log.Println(source, k, v.Hash, v.Other)
		// v.Hash = hash(v)
		kv := make(map[string]string)
		switch source {
		case "original", "target":
			kv["hash"] = v.Hash
		}
		if v.Zero != "" {
			kv["zero"] = v.Zero
		}
		if v.One != "" {
			kv["one"] = v.One
		}
		if v.Two != "" {
			kv["two"] = v.Two
		}
		if v.Few != "" {
			kv["few"] = v.Few
		}
		if v.Many != "" {
			kv["many"] = v.Many
		}
		if v.Other != "" {
			kv["other"] = v.Other
		}
		t[k] = kv
	}

	out, err := yaml.Marshal(t)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, out, os.ModePerm)
}

func init() {
	logrus.SetLevel(logrus.TraceLevel)

	cobra.OnInitialize(readConfigFile)
}

func main() {
	var rootCmd = &cobra.Command{Use: "goi18n"}

	// global flags

	// verbose
	rootCmd.PersistentFlags().Bool("verbose", false, "啰嗦模式")
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))

	// default
	rootCmd.PersistentFlags().StringP("default", "d", "en-US", "默认语言")
	rootCmd.MarkFlagRequired("default")
	viper.BindPFlag("default", rootCmd.PersistentFlags().Lookup("default"))

	// outdir
	rootCmd.Flags().String("outdir", "locales", "输出文件夹")
	viper.BindPFlag("outdir", rootCmd.Flags().Lookup("outdir"))

	// sub command walk
	{
		cmd := &cobra.Command{
			Use:              "walk",
			Short:            "遍历项目文件夹找出翻译语句",
			Long:             "",
			TraverseChildren: true,
			// Args:  cobra.MinimumNArgs(1), // 至少需要几个参数
			Run: func(cmd *cobra.Command, args []string) {
				walk()
			},
		}
		cmd.Flags().StringArray("path", nil, "")
		viper.BindPFlag("path", cmd.Flags().Lookup("path"))

		cmd.Flags().Bool("ignore-test-files", true, "是否忽略 _test.go 文件")
		viper.BindPFlag("ignore-test-files", cmd.Flags().Lookup("ignore-test-files"))

		rootCmd.AddCommand(cmd)
	}

	// sub command merge
	{
		cmd := &cobra.Command{
			Use:              "merge",
			Short:            "遍历项目文件夹找出翻译语句",
			Long:             "",
			TraverseChildren: true,
			// Args:             cobra.MinimumNArgs(1), // 至少需要1个参数
			Run: func(cmd *cobra.Command, args []string) {
				merge()
			},
		}

		// target
		cmd.PersistentFlags().StringArrayP("target", "t", nil, "目标语言")
		cmd.MarkFlagRequired("target")
		viper.BindPFlag("target", cmd.PersistentFlags().Lookup("target"))

		rootCmd.AddCommand(cmd)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// writeConfigFile()
}

// 从 .goi18n.yaml 文件 读取配置
func readConfigFile() {
	gomod := getGoEnv("GOMOD")
	if gomod == "" {
		return
	}
	workspace := filepath.Dir(gomod)

	// 读取配置
	viper.AddConfigPath(workspace)
	viper.SetConfigFile(configFileName)
	viper.SetConfigType(configFileType)

	// 忽略错误
	_ = viper.ReadInConfig()
}

// 在 go mod 模式的项目下 自动生成 .goi18n.yaml 文件
func writeConfigFile() {
	gomod := getGoEnv("GOMOD")
	if gomod == "" {
		return
	}
	workspace := filepath.Dir(gomod)

	// 创建配置文件（使用默认配置）
	tmp := viper.New()
	tmp.AddConfigPath(workspace)
	tmp.SetConfigFile(configFileName)
	tmp.SetConfigType(configFileType)
	tmp.Set("default", viper.Get("default"))
	tmp.Set("target", viper.GetStringSlice("target"))
	err := tmp.WriteConfig()
	if err != nil {
		logrus.Errorln(err)
	}
}

var goEnvCache = make(map[string]string)

func getGoEnv(key string) string {
	if val, ok := goEnvCache[key]; ok {
		return val
	}
	out, err := exec.Command("go", "env", key).Output()
	if err != nil {
		panic(err.Error())
	}
	goEnvCache[key] = string(out)
	return string(out)
}
