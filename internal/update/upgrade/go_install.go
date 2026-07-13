package upgrade

import (
	"os"
	"path/filepath"
	"strings"
)

var (
	currentExecutable = os.Executable
	goInstallBinDirs  = defaultGoInstallBinDirs
)

// goInstallOwnsActiveExecutableForOS reports whether the running binary is Go's
// direct canonical install target. Go has no install provenance manifest, so
// `go env GOBIN GOPATH` is the authoritative ownership boundary.
func goInstallOwnsActiveExecutableForOS(osName string) bool {
	activePath, err := currentExecutable()
	if err != nil || activePath == "" {
		return false
	}

	for _, binDir := range goInstallBinDirs() {
		if goInstallTargetMatches(activePath, binDir, osName) {
			return true
		}
	}

	return false
}

func goInstallTargetMatches(activePath, binDir, osName string) bool {
	target := filepath.Join(binDir, gentleAIExecutableName(osName))
	resolvedActivePath, err := filepath.EvalSymlinks(activePath)
	if err == nil {
		activePath = resolvedActivePath
	}
	resolvedTarget, err := filepath.EvalSymlinks(target)
	if err == nil {
		target = resolvedTarget
	}

	activePath = filepath.Clean(activePath)
	target = filepath.Clean(target)
	if osName == "windows" {
		return strings.EqualFold(activePath, target)
	}
	return activePath == target
}

func gentleAIExecutableName(osName string) string {
	if osName == "windows" {
		return "gentle-ai.exe"
	}
	return "gentle-ai"
}

func defaultGoInstallBinDirs() []string {
	out, err := execCommand("go", "env", "GOBIN", "GOPATH").Output()
	if err != nil {
		return nil
	}
	return goInstallBinDirsFromEnv(string(out))
}

func goInstallBinDirsFromEnv(output string) []string {
	values := strings.Split(output, "\n")
	if len(values) == 0 {
		return nil
	}

	if gobin := strings.TrimSpace(values[0]); gobin != "" {
		return []string{gobin}
	}
	if len(values) < 2 {
		return nil
	}

	var binDirs []string
	for _, goPath := range filepath.SplitList(strings.TrimSpace(values[1])) {
		if goPath != "" {
			binDirs = append(binDirs, filepath.Join(goPath, "bin"))
		}
	}
	return binDirs
}
