package npm

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	SmithyBuildPath = "smithy/build/smithyprojections/smithy/source/typescript-ssdk-codegen"
)

// Link runs `npm link` in the given directory to register a package globally
func Link(dir string) error {
	cmd := exec.Command("npm", "link")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// LinkPackage runs `npm link <pkg>` in the given directory to consume a linked package
func LinkPackage(dir, pkg string) error {
	cmd := exec.Command("npm", "link", pkg)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Unlink runs `npm unlink <pkg>` in the given directory
func Unlink(dir, pkg string) error {
	cmd := exec.Command("npm", "unlink", pkg)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// IsBuilt checks if a Smithy model directory has built artifacts
func IsBuilt(modelDir string) bool {
	buildDir := filepath.Join(modelDir, SmithyBuildPath)

	packageJSON := filepath.Join(buildDir, "package.json")
	distTypes := filepath.Join(buildDir, "dist-types")

	if _, err := os.Stat(packageJSON); err != nil {
		return false
	}
	if _, err := os.Stat(distTypes); err != nil {
		return false
	}
	return true
}

// BuildOutputDir returns the path to the Smithy SDK build output
func BuildOutputDir(modelDir string) string {
	return filepath.Join(modelDir, SmithyBuildPath)
}

// GetPackageName reads the package name from a package.json file
func GetPackageName(dir string) (string, error) {
	packageJSON := filepath.Join(dir, "package.json")
	if _, err := os.Stat(packageJSON); err != nil {
		return "", fmt.Errorf("package.json not found in %s", dir)
	}

	cmd := exec.Command("node", "-p", "require('./package.json').name")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to read package name: %w", err)
	}

	name := string(out)
	if len(name) > 0 && name[len(name)-1] == '\n' {
		name = name[:len(name)-1]
	}
	return name, nil
}

// IsLinked checks if a package is currently npm-linked in the given directory
func IsLinked(dir, pkg string) bool {
	nodeModulesPath := filepath.Join(dir, "node_modules", pkg)
	info, err := os.Lstat(nodeModulesPath)
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeSymlink != 0
}

// CheckNPM verifies that npm is installed
func CheckNPM() error {
	_, err := exec.LookPath("npm")
	if err != nil {
		return fmt.Errorf("npm not found — install Node.js from https://nodejs.org")
	}
	return nil
}
