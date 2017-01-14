package main

import (
	"os"
	"path"
	"path/filepath"
	"fmt"
	"io"
	"io/ioutil"
	"github.com/urfave/cli"
	"os/exec"
)

func main() {
	app := cli.NewApp()
	app.Name = "WorkshopGitBranches"
	app.Version = "1.0"
	app.HelpName = "wgb"
	app.Usage = "create GIT repositories with branches for each assignment in a workshop"
	app.Authors = []cli.Author{
		cli.Author{
			Name:"Joost van Wollingen",
			Email:"joostvanwollingen@gmail.com",
		},
	}

	app.Commands = []cli.Command{
		{
			Name: "init",
			Usage: "create basic folder structure, including branches and shared folders",
			Action: func(c *cli.Context) error {
				return doInit(c)
			},
			Flags:  []cli.Flag{
				cli.BoolTFlag{
					Name:        "shared-dir",
					Usage:       "boolean value to toggle the creation of the shared directory",
				}},
		},
		{
			Name: "assemble",
			Usage: "assemble all files needed for the workshop.",
			ArgsUsage: "[workshop directory]",
			Action: func(c *cli.Context) error {
				return assembleBranches(c)

			},
		},
		{
			Name: "build",
			Usage: "assemble and create branches",
			Action: func(c *cli.Context) error {
				err := assembleBranches(c)
				if (err == nil) {
					err = createBranches(c)
				}
				return err
			},
		},
	}
	app.Run(os.Args)
}
func createBranches(c *cli.Context) error {
	assemblyDir := "target"
	branches := getDirectories(assemblyDir)

	if (len(branches) > 0) {

		for _, branch := range branches {
			info(branch.Name())
			executeCommand("git","branch","-D",branch.Name())
			executeCommand("git","checkout","--orphan",branch.Name())
			executeCommand("git","rm","-rf",".")
			CopyDir(path.Join(assemblyDir,branch.Name()), ".")
			executeCommand("git","add","--all")
			executeCommand("git","reset",assemblyDir)
			executeCommand("git","commit","-am","Created branch")
			executeCommand("git","checkout","master")
		}
	} else {
		return cli.NewExitError(fmt.Sprintf("No branches found in %s", assemblyDir), 1)
	}
	return nil
}
func info(message string) {
	fmt.Println(message)

}

func executeCommand(command string,arguments...string){
	exec.Command(command, arguments...).Output()
}

func assembleBranches(c *cli.Context) error {
	sourceDir := defaultOrFirstArg(".", c)
	sharedDir := path.Join(sourceDir, "shared")

	targetDirectory := createTargetDirectory()

	branches := getDirectories(path.Join(sourceDir, "branches"))

	if (len(branches) > 0) {
		for _, branch := range branches {
			CopyDir(sharedDir, path.Join(targetDirectory, branch.Name()))
			CopyDir(path.Join(sourceDir, "branches", branch.Name()), path.Join(targetDirectory, branch.Name()))
		}
	} else {
		return cli.NewExitError(fmt.Sprintf("No branches found in %s", sourceDir), 1)
	}
	info(fmt.Sprintf("Assembled %d branches in %s", len(branches), targetDirectory))
	return nil
}
func createTargetDirectory() string {
	targetDirectory := "target"
	removeDirectory(targetDirectory)
	createDirectory(targetDirectory)
	return targetDirectory
}
func removeDirectory(folder string) {
	os.RemoveAll(folder)
}

func getDirectories(folder string) []os.FileInfo {
	directories := []os.FileInfo{}
	files, _ := ioutil.ReadDir(folder)
	for _, file := range files {
		if (file.IsDir()) {
			directories = append(directories, file)
		}
	}
	return directories
}

func doInit(c *cli.Context) error {
	directory := defaultOrFirstArg(".", c)
	initDirectory(directory)
	if (c.BoolT("shared-dir")) {
		createDirectory(path.Join(directory, "shared"))
	}
	info(fmt.Sprintf("Initialized new workshop in %s", directory))
	return nil
}

func initDirectory(directory string) {
	createDirectory(directory)
	createDirectory(path.Join(directory, "branches"))
}

func createDirectory(directory string) {
	err := os.MkdirAll(directory, os.ModePerm)
	if err != nil {
		fmt.Println("MkdirAll %q: %s", directory, err)
	}
}

func defaultOrFirstArg(defaultValue string, firstArgument *cli.Context) string {
	value := defaultValue

	if (firstArgument.Args().Present()) {
		value = firstArgument.Args().First();
	}
	return value
}

// CopyFile copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file. The file mode will be copied from the source and
// the copied data is synced/flushed to stable storage.
func CopyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		if e := out.Close(); e != nil {
			err = e
		}
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return
	}

	err = out.Sync()
	if err != nil {
		return
	}

	si, err := os.Stat(src)
	if err != nil {
		return
	}
	err = os.Chmod(dst, si.Mode())
	if err != nil {
		return
	}

	return
}

// CopyDir recursively copies a directory tree, attempting to preserve permissions.
// Source directory must exist, destination directory must *not* exist.
// Symlinks are ignored and skipped.
func CopyDir(src string, dst string) (err error) {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !si.IsDir() {
		return fmt.Errorf("source is not a directory")
	}

	err = os.MkdirAll(dst, si.Mode())
	if err != nil {
		return
	}

	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = CopyDir(srcPath, dstPath)
			if err != nil {
				return
			}
		} else {
			// Skip symlinks.
			if entry.Mode() & os.ModeSymlink != 0 {
				continue
			}

			err = CopyFile(srcPath, dstPath)
			if err != nil {
				return
			}
		}
	}

	return
}
