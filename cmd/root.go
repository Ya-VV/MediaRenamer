package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yavitvas/yaRenamer/pkg"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "yaRenamer",
	Short: "Utility for batch processing files in a directory, resulting in one pattern: YearMonthDayofmonth_HourMinuteSecond",
	Long: `Utility for batch processing files in a directory.
	Renames all supported formats based on creation date, resulting in one pattern:
	YearMonthDayofmonth_HourMinuteSecond (example: 20201009_115900).`,
	Run: func(cmd *cobra.Command, args []string) {
		pkg.LetsGo()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.Flags().BoolP("verbose", "v", false, "set verbose output")
	rootCmd.Flags().Bool("check-dublicates", false, "to check files dublicates")
	rootCmd.Flags().String("dir", "", "Put the path to directory")
}

func initConfig() {
	if verboseStatus, err := rootCmd.Flags().GetBool("verbose"); err == nil {
		pkg.SetVerbose(verboseStatus)
	}
	if checkDubleStatus, err := rootCmd.Flags().GetBool("check-dublicates"); err == nil {
		pkg.SetCheckDublesFlag(checkDubleStatus)
	}
	pkg.SetWorkDir(rootCmd.Flags().GetString("dir"))
}
