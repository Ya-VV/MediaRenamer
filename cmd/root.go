package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yavitvas/yaRenamer/pkg"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "yaRenamer",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		pkg.LetsGo()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.snike.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("verbose", "v", false, "set verbose output")
	rootCmd.Flags().Bool("check-dublicates", false, "to check files dublicates")
	rootCmd.Flags().String("dir", "", "Put the path to directory")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if verboseStatus, err := rootCmd.Flags().GetBool("verbose"); err == nil {
		pkg.SetVerbose(verboseStatus)
	}
	if checkDubleStatus, err := rootCmd.Flags().GetBool("check-dublicates"); err == nil {
		pkg.SetCheckDublesFlag(checkDubleStatus)
	}
	pkg.SetWorkDir(rootCmd.Flags().GetString("dir"))
	// if cfgFile != "" {
	// 	// Use config file from the flag.
	// 	viper.SetConfigFile(cfgFile)
	// } else {
	// 	// Find home directory.
	// 	home, err := homedir.Dir()
	// 	if err != nil {
	// 		fmt.Println(err)
	// 		os.Exit(1)
	// 	}

	// 	// Search config in home directory with name ".snike" (without extension).
	// 	viper.AddConfigPath(home)
	// 	viper.SetConfigName(".snike")
	// }

	// viper.AutomaticEnv() // read in environment variables that match

	// // If a config file is found, read it in.
	// if err := viper.ReadInConfig(); err == nil {
	// 	fmt.Println("Using config file:", viper.ConfigFileUsed())
	// }
}
