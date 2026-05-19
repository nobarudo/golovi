package cmd

import (
	"fmt"
	"os"

	"golovi/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "golovi [file]",
	Short: "A brief description of your application",
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var logFile string
		if len(args) > 0 {
			logFile = args[0]
		} else {
			// デフォルトで log.text を見るか、エラーを出すか
			// ここでは log.text があればそれを使うようにしてみる
			if _, err := os.Stat("log.text"); err == nil {
				logFile = "log.text"
			} else {
				return fmt.Errorf("log file required")
			}
		}

		m, err := tui.NewModel(logFile)
		if err != nil {
			return err
		}

		p := tea.NewProgram(m, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			return err
		}
		return nil
	},
	}

	// Execute adds all child commands to the root command and sets flags appropriately.
	// This is called by main.main(). It only needs to happen once to the rootCmd.
	func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
	}

	func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	}