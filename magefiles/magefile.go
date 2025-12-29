//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/magefile/mage/mg" // mg contains helpful utility functions, like Deps
)

// Default target to run when none is specified
// If not set, running mage will list available targets
var Default = Run

// A build step that requires additional params, or platform specific steps for example
func Build() error {
	mg.Deps(InstallDeps)
	fmt.Println("Building...")
	cmd := exec.Command("go", "build", "-o", "MyApp", ".")
	return cmd.Run()
}

// A custom install step if you need your bin someplace other than go/bin
func Install() error {
	mg.Deps(Build)
	fmt.Println("Installing...")
	return os.Rename("./MyApp", "/usr/bin/MyApp")
}

// Runs templ generator
func Templ() error {
	fmt.Println("Running templ generator...")
	cmd := exec.Command("templ", "generate")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Runs the app
func Run() error {
	mg.Deps(InstallDeps)
	fmt.Println("Runing...")
	cmd := exec.Command("go", "run", "main.go")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Manage your deps, or running package managers.
func InstallDeps() error {
	fmt.Println("Running templ generator...")
	cmd := exec.Command("templ", "generate")
	return cmd.Run()
}

// Clean up after yourself
func Clean() {
	fmt.Println("Cleaning...")
	os.RemoveAll("MyApp")
}
