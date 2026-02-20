package npm

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	SmithyBuildBase = "smithy/build/smithyprojections/smithy/source"
	// Default codegen for server SDKs
	SmithyBuildPath = SmithyBuildBase + "/typescript-ssdk-codegen"
)

// DirectLink creates a symlink from consumerDir/node_modules/<pkg> -> buildDir.
// No npm commands are invoked, so no registry auth is needed.
func DirectLink(consumerDir, pkg, buildDir string) error {
	target := filepath.Join(consumerDir, "node_modules", pkg)

	// Ensure the parent scope directory exists (e.g. node_modules/@spark-rewards)
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(target), err)
	}

	// Remove whatever is there (regular dir or old symlink)
	if err := os.RemoveAll(target); err != nil {
		return fmt.Errorf("remove %s: %w", target, err)
	}

	absBuild, err := filepath.Abs(buildDir)
	if err != nil {
		return err
	}

	return os.Symlink(absBuild, target)
}

// Unlink removes a symlinked package and does NOT reinstall the published
// version — the next `npm install` (or spark-cli sync) will restore it.
func Unlink(consumerDir, pkg string) error {
	target := filepath.Join(consumerDir, "node_modules", pkg)
	info, err := os.Lstat(target)
	if err != nil {
		return nil // nothing to unlink
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return nil // not a symlink, leave it alone
	}
	return os.Remove(target)
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

// BuildOutputDir returns the path to the default Smithy SDK build output
func BuildOutputDir(modelDir string) string {
	return filepath.Join(modelDir, SmithyBuildPath)
}

// BuildOutputDirForCodegen returns the path for a specific codegen type
func BuildOutputDirForCodegen(modelDir, codegen string) string {
	return filepath.Join(modelDir, SmithyBuildBase, codegen)
}

// IsBuiltForCodegen checks if a specific codegen output exists
func IsBuiltForCodegen(modelDir, codegen string) bool {
	buildDir := BuildOutputDirForCodegen(modelDir, codegen)
	packageJSON := filepath.Join(buildDir, "package.json")
	if _, err := os.Stat(packageJSON); err != nil {
		return false
	}
	return true
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
