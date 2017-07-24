# workshopgitbranches

Workshops always have that little extra if there is an actual codebase available that people can use as a start, of of which the exercises can build, etc. etc. Often, these GIT repo's will have exercises and example solutions on separate branches. Maintaining a repository like that is a pain in the ass, especially when you have to update a shared file on 16 different branches. 

This little commandline tool, written in Golang, simplifies that a bit for you. 

NOTE: There are no tests for the code (yet), so use at your own risk.

## Building the code

1. Run `go get github.com/joostvanwollingen/workshopgitbranches`.
2. Run `go get -d ./...` to download the dependencies (only http://github.com/urfave/cli at the moment)
3. Run `go install` in the workshopgitbranches folder
4. Run `workshopgitbranches -v` to check if everything is setup correctly.

Or just grab a binary from the [releases](https://github.com/joostvanwollingen/workshopgitbranches/releases).

## Usage

*__workshopgitbranches assumes you do every operation on master for now. Don't use it from other branches.__*

### First time use
To create the `branches` and `shared` folders:

1. Ensure you are on master.
2. Run `workshopgitbranches init`.

### Building branches
1. Put files that you need on every branch in the `shared` folder.
2. Put branch specific files in the `branches\<branchname>` folder.
3. Run `workshopgitbranches assemble` to test your build. This will create a `target` folder containing folders for each branch, merged with the shared files.
Note: files that are both in the shared and branch specific folders, will be overridden with the branch specific version.
4. Running `workshopgitbranches build` will: **delete all branches in the GIT repository** that are in your branches folder, merge the shared and branch folders, create one commit with these files on each branch.
 
### Example
 
Running `workshopgitbranches` against the following example directory structure would result in 5 branches, named `1_first`, `1_first_solution`, `2_second`, `3_third`, `anotherbranch`. Each of the branches will contain all the files below the branch folder and the files from the shared folder. All files are included recursively.
 
    myworkshop/ (on master branch)
    ├── branches
    │   ├── 1_first
    │   │   └── 1.txt
    │   ├── 1_first_solution
    │   │   └── 1.txt
    │   ├── 2_second
    │   │   ├── 2.txt
    │   │   └── src
    │   │       └── main
    │   │           └── java
    │   │               └── Main.java
    │   ├── 3_third
    │   │   └── 3.txt
    │   ├── anotherbranch
    │   │   └── another.txt
    │   └── ignore.file
    └── shared
        ├── shared
        │   └── gotcha
        ├── 1.txt
        ├── readme.md
        └── we need this too.file

