package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-tools/gows/gows"

	"gopkg.in/viktorbenei/cobra.v0"
)

var cfgFile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "goinst go/package/name",
	Short: "Go install command line tools, in an isolated Go workspace",
	Long:  ``,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	RootCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("No command specified!")
		}
		RootCmd.SilenceErrors = true
		RootCmd.SilenceUsage = true
		return nil
	}

	RootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		cmdName := args[0]
		if cmdName == "-h" || cmdName == "--help" {
			if err := RootCmd.Help(); err != nil {
				return err
			}
			return nil
		}

		packageToInstall := args[0]
		return goInstallPackageInIsolation(packageToInstall)
	}
}

func goInstallPackageInIsolation(packageToInstall string) error {
	fmt.Println("=> Installing package " + packageToInstall + " ...")
	workspaceRootPath, err := pathutil.NormalizedOSTempDirPath("goinst")
	if err != nil {
		return fmt.Errorf("Failed to create root directory of isolated workspace, error: %s", err)
	}
	fmt.Println("=> Using sandboxed workspace:", workspaceRootPath)

	origGOPATH := os.Getenv("GOPATH")
	if origGOPATH == "" {
		return fmt.Errorf("You don't have a GOPATH environment - please set it; GOPATH/bin will be symlinked")
	}

	fmt.Println("=> Symlink GOPATH/bin into sandbox ...")
	if err := gows.CreateGopathBinSymlink(origGOPATH, workspaceRootPath); err != nil {
		return fmt.Errorf("Failed to create GOPATH/bin symlink, error: %s", err)
	}
	fmt.Println("   [DONE]")

	fmt.Println("=> Installing package " + packageToInstall + " ...")
	cmd := gows.CreateCommand(workspaceRootPath, workspaceRootPath, "go", "get", "-u", "-v", packageToInstall)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Failed to install package, error: %s", err)
	}
	fmt.Println("   [DONE] Package successfully installed")

	fmt.Println("=> Delete isolated workspace ...")
	if err := os.RemoveAll(workspaceRootPath); err != nil {
		return fmt.Errorf("Failed to delete temporary isolated workspace, error: %s", err)
	}
	fmt.Println("   [DONE]")

	return nil
}
