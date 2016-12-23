// Copyright © 2016 Mike Hudgins <mchudgins@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"fmt"

	"github.com/mchudgins/certMgr/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// debugCmd represents the debug command
var debugCmd = &cobra.Command{
	Hidden: true,
	Use:    "debug",
	Short:  "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("debug called")
		//		cmd.Flags().VisitAll(displayCobraFlag)

		cmd.Flags().VisitAll(displayFlagInfo)

		viper.BindPFlags(cmd.Flags())
		viper.SetEnvPrefix("certMgr")
		viper.AutomaticEnv()
		fmt.Printf("\nAutomaticEnv was called.\n")
		cmd.Flags().VisitAll(displayFlagInfo)

		viper.SetConfigName("config")
		viper.AddConfigPath("$HOME/.certMgr")
		err := viper.ReadInConfig()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "%s\n", err)
		}

		fmt.Printf("\nReadInConfig was called.\n")
		cmd.Flags().VisitAll(displayFlagInfo)

		fmt.Printf("verbose: %t\n", viper.GetBool("verbose"))
		fmt.Printf("int:  %d\n", viper.GetInt("int"))
		fmt.Printf("str:  %s\n", viper.GetString("str"))

		cfg := &rootConfig{Debug: defaultDebugConfig}
		cfg.Debug = defaultDebugConfig

		err = viper.Unmarshal(cfg)

		//viper.Unmarshal(&cfg.Debug)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "unmarshal -- %s", err)
		}

		fmt.Printf("cfg: %+v\n", cfg)
	},
}

var debug2Cmd = &cobra.Command{
	Hidden: true,
	Use:    "debug2",
	Short:  "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		// determine the backend app's configuration
		cfg := &rootConfig{Debug: defaultDebugConfig}

		err := utils.NewConfig(cmd, defaultConfig, cfg)
		if err != nil {
			fmt.Printf("an error occurred while obtaining the application configuration -- %s", err)
		}

		if cfg.Verbose {
			fmt.Printf("Verbose logging enabled.\n")
		}

		cfg.Debug.Flag = viper.GetBool("flag")
		//		cfg.Debug.AnInteger = viper.GetInt("int")

		fmt.Printf("cfg: %+v\n", cfg)
	},
}

type rootConfig struct {
	Config  string
	Verbose bool
	Debug   debugConfig
}

type debugConfig struct {
	Flag      bool   `json:"flag"`
	AnInteger int    `json:"int"`
	Str       string `json:"str"`
	NoChange  string
}

var (
	defaultDebugConfig = debugConfig{NoChange: "This string should be const", AnInteger: 123}
)

func displayFlagInfo(f *pflag.Flag) {
	var set string
	var inConfig string

	if viper.IsSet(f.Name) {
		set = ""
	} else {
		set = "NOT "
	}

	if viper.InConfig(f.Name) {
		inConfig = ""
	} else {
		inConfig = "NOT "
	}
	fmt.Printf("%s was %sset; was %sin config.\n", f.Name, set, inConfig)
}

func init() {
	RootCmd.AddCommand(debugCmd)
	RootCmd.AddCommand(debug2Cmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// debugCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// debugCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	debugCmd.PersistentFlags().Bool("flag", defaultDebugConfig.Flag, "a boolean flag")
	debug2Cmd.PersistentFlags().Bool("flag", defaultDebugConfig.Flag, "a boolean flag")
	debugCmd.PersistentFlags().Int("int", defaultDebugConfig.AnInteger, "an integer flag")
	debug2Cmd.PersistentFlags().Int("int", defaultDebugConfig.AnInteger, "an integer flag")
	debugCmd.PersistentFlags().String("str", defaultDebugConfig.Str, "a string flag")
}
