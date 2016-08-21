package cmd

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/bitrise-io/go-utils/fileutil"
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

func copyFile(fromPath, toPath string) (copyErr error) {
	srcFilePerm, err := fileutil.GetFilePermissions(fromPath)
	if err != nil {
		return fmt.Errorf("Failed to get file permission of source, error: %s", err)
	}

	in, err := os.Open(fromPath)
	if err != nil {
		return fmt.Errorf("Failed to open source for read (%s), error: %s", fromPath, err)
	}
	defer func() {
		if err := in.Close(); err != nil {
			if copyErr == nil {
				copyErr = fmt.Errorf("Failed to close source, error: %s", err)
			}
		}
	}()
	out, err := os.Create(toPath)
	if err != nil {
		return fmt.Errorf("Failed to create target for write (%s), error: %s", toPath, err)
	}
	defer func() {
		if err := out.Close(); err != nil {
			if copyErr == nil {
				copyErr = fmt.Errorf("Failed to close target, error: %s", err)
			}
		}
	}()
	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("Failed to copy content, error: %s", err)
	}
	if err := out.Sync(); err != nil {
		return fmt.Errorf("Failed to sync output, error: %s", err)
	}

	return os.Chmod(toPath, srcFilePerm)
}

func copyFilesFromDir(srcDir, targetDir string) error {
	files, err := ioutil.ReadDir(srcDir)
	if err != nil {
		return fmt.Errorf("Failed to read source directory content, error: %s", err)
	}

	for _, file := range files {
		srcFilePath := filepath.Join(srcDir, file.Name())
		destFilePath := filepath.Join(targetDir, file.Name())
		fmt.Println("   * " + srcFilePath + " -> " + destFilePath)
		if err := copyFile(srcFilePath, destFilePath); err != nil {
			return fmt.Errorf("Failed to copy file, error: %s", err)
		}
	}

	return nil
}

func goInstallPackageInIsolation(packageToInstall string) error {
	fmt.Println("=> Installing package " + packageToInstall + " ...")
	workspaceRootPath, err := pathutil.NormalizedOSTempDirPath("goinst")
	if err != nil {
		return fmt.Errorf("Failed to create root directory of isolated workspace, error: %s", err)
	}
	fmt.Println("=> Using sandboxed workspace:", workspaceRootPath)

	// origGOPATH := os.Getenv("GOPATH")
	// if origGOPATH == "" {
	// 	return fmt.Errorf("You don't have a GOPATH environment - please set it; GOPATH/bin will be symlinked")
	// }

	// fmt.Println("=> Symlink GOPATH/bin into sandbox ...")
	// if err := gows.CreateGopathBinSymlink(origGOPATH, workspaceRootPath); err != nil {
	// 	return fmt.Errorf("Failed to create GOPATH/bin symlink, error: %s", err)
	// }
	// fmt.Println("   [DONE]")

	fmt.Println("=> Installing package " + packageToInstall + " ...")
	{
		cmd := gows.CreateCommand(workspaceRootPath, workspaceRootPath,
			"go", "get", "-u", "-v", packageToInstall)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("Failed to install package, error: %s", err)
		}
	}
	fmt.Println("   [DONE] Package successfully installed")

	binTargetDirPath := "/usr/local/bin"
	fmt.Println("=> Copying generated binaries into " + binTargetDirPath + " ...")
	{
		gopathBinPath := filepath.Join(workspaceRootPath, "bin")
		if err := copyFilesFromDir(gopathBinPath, binTargetDirPath); err != nil {
			return fmt.Errorf("Failed to move binaries, error: %s", err)
		}
	}
	fmt.Println("   [DONE]")

	fmt.Println("=> Delete isolated workspace ...")
	{
		if err := os.RemoveAll(workspaceRootPath); err != nil {
			return fmt.Errorf("Failed to delete temporary isolated workspace, error: %s", err)
		}
	}
	fmt.Println("   [DONE]")

	return nil
}
