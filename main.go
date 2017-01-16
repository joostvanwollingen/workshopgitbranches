//Easily maintain workshop Git repositories by automatically creating branches for each assignment
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
	app.HelpName = "workshopgitbranches"
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

//Initialize a new workshop directory with optional shared directory
func doInit(c *cli.Context) error {
	directory := defaultOrFirstArg(".", c)
	createDirectory(path.Join(directory, "branches"))

	if (c.BoolT("shared-dir")) {
		createDirectory(path.Join(directory, "shared"))
	}

	info(fmt.Sprintf("Initialized new workshop in %s", directory))
	return nil
}

//Create a target directory based on the branches from the current directory or argument passed
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

//Create branches based on the folders in ./target
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

//Exec.command with arguments, don't bother with output
func executeCommand(command string,arguments...string){
	exec.Command(command, arguments...).Output()
}

//Cleans and recreates the target directory
func createTargetDirectory() string {
	targetDirectory := "target"
	removeDirectory(targetDirectory)
	createDirectory(targetDirectory)
	return targetDirectory
}

//Remove a directory recursively
func removeDirectory(folder string) {
	os.RemoveAll(folder)
}

//Get a slide of os.FileInfo of all directories inside a directory
//Ignores files
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

//Create a directory
func createDirectory(directory string) {
	err := os.MkdirAll(directory, os.ModePerm)
	if err != nil {
		fmt.Println("MkdirAll %q: %s", directory, err)
	}
}

//Get the defaultValue or the first argument passed on cli
func defaultOrFirstArg(defaultValue string, firstArgument *cli.Context) string {
	value := defaultValue

	if (firstArgument.Args().Present()) {
		value = firstArgument.Args().First();
	}
	return value
}

// Retrieved from https://gist.github.com/m4ng0squ4sh/92462b38df26839a3ca324697c8cba04
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

// Retrieved from https://gist.github.com/m4ng0squ4sh/92462b38df26839a3ca324697c8cba04
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
