package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/umlx5h/zsh-manpage-completion-generator/internal/converter"
	"github.com/umlx5h/zsh-manpage-completion-generator/internal/util"
)

// These variables are set in build step
var (
	version = "unset"
	commit  = "unset"
	date    = "unset"
)

var (
	excludeCmds = []string{
		"[",
		"]",
		"sudo",
	}
)

func main() {
	dataDir := os.Getenv("XDG_DATA_HOME")
	if dataDir == "" {
		dataDir = filepath.Join(os.Getenv("HOME"), ".local/share")
	}
	filepath.Join(dataDir)
	srcDir := filepath.Join(dataDir, "fish/generated_completions")
	dstDir := filepath.Join(dataDir, "zsh/generated_man_completions")

	src := flag.String("src", srcDir, "fish generated_completions src folder")
	dst := flag.String("dst", dstDir, "zsh generated_completions destination folder")
	clean := flag.Bool("clean", false, "CAUTION: remove destination folder before converting")
	verbose := flag.Bool("verbose", false, "verbose log")
	version_ := flag.Bool("version", false, "show version")
	flag.Parse()

	srcDir = *src
	dstDir = *dst
	isClean := *clean
	isVerbose := *verbose
	isVersion := *version_

	if isVersion {
		if commit != "unset" {
			fmt.Printf("Version: %s, Commit: %s, Date: %s\n", version, commit, date)
		} else if buildInfo, ok := debug.ReadBuildInfo(); ok {
			fmt.Printf("Version: %s\n", buildInfo.Main.Version)
		} else {
			fmt.Printf("Version: %s\n", "(unknown)")
		}

		os.Exit(0)
	}

	srcDirEntry, err := os.ReadDir(srcDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not open srcDir, are you sure to install fish?: %s", err)
		os.Exit(1)
	}
	dir, err := os.Stat(dstDir)
	if err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "could not open dstDir: %s", err)
		os.Exit(1)
	} else if err == nil && !dir.IsDir() {
		fmt.Fprintf(os.Stderr, "could not open dstDir as directory: %s", dstDir)
		os.Exit(1)
	}

	if isClean {
		fmt.Printf("Cleaning zsh completions folder: %s\n", dstDir)
		if err = os.RemoveAll(dstDir); err != nil {
			fmt.Fprintf(os.Stderr, "could not remove dstDir: %s", err)
			os.Exit(1)
		}

		if err = os.MkdirAll(dstDir, 0777); err != nil {
			fmt.Fprintf(os.Stderr, "could not mkdir dstDir: %s", err)
			os.Exit(1)
		}
	} else if os.IsNotExist(err) {
		if err = os.MkdirAll(dstDir, 0777); err != nil {
			fmt.Fprintf(os.Stderr, "could not mkdir dstDir: %s", err)
			os.Exit(1)
		}
	}

	var (
		// stat
		convertNum   int
		convertedNum int
		skippedNum   int
	)

	fmt.Printf("Converting fish completions: %s -> %s\n", srcDir, dstDir)

	for _, f := range srcDirEntry {
		func() {
			if f.Type().IsRegular() {
				fileName := f.Name()
				if !strings.HasSuffix(fileName, ".fish") {
					fmt.Printf("skipped non fish file: %s\n", fileName)
					return
				}
				if strings.HasSuffix(fileName, ".1posix.fish") {
					// In zsh, posix and non-posix versions need to be represented as one common command, so skip posix version
					// TODO: if only found posix version, then convert it.
					if isVerbose {
						fmt.Printf("skipped posix version: %s\n", fileName)
					}
					skippedNum++
					return
				}

				cmdName := strings.TrimSuffix(fileName, ".fish")

				// delete unneeded command
				if util.Contains(excludeCmds, cmdName) {
					if isVerbose {
						fmt.Printf("skipped unneeded command: %s\n", fileName)
					}
					skippedNum++
					return
				}

				srcFilePath := filepath.Join(srcDir, fileName)
				srcFile, err := os.Open(filepath.Join(srcFilePath))
				if err != nil {
					fmt.Fprintf(os.Stderr, "could not open srcFile: %s: %s", srcFilePath, err)
					os.Exit(1)
				}
				defer srcFile.Close()
				convertNum++
				converter := converter.NewConverter(srcFile, cmdName)
				fileContent, err := converter.Convert()
				if err != nil {
					if isVerbose {
						fmt.Printf("failed to convert: %s: %s\n", fileName, err)
					}
					return
				}

				dstFilePath := filepath.Join(dstDir, "_"+cmdName)
				dstFile, err := os.Create(dstFilePath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "could not create dstFile: %s: %s", dstFilePath, err)
					os.Exit(1)
				}
				defer dstFile.Close()

				dstFile.WriteString(fileContent)

				if isVerbose {
					fmt.Printf("converted: %s\n", f.Name())
				}

				convertedNum++
			}
		}()
	}

	fmt.Printf("Completed. converted: %d/%d, skipped: %d\n", convertedNum, convertNum, skippedNum)
}
