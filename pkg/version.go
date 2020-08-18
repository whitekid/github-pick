package pocket

import (
	"fmt"
	"runtime"
)

var (
	GitCommit = ""
	GitSHA    = ""
	GitBranch = ""
	GitTag    = ""
	GitDirty  = ""
	BuildTime = ""
)

// Version print version informations
func Version() {
	fmt.Printf("Go Version: %s\n", runtime.Version())
	fmt.Printf("Compiler: %s\n", runtime.Compiler)
	fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("Git commit: %s\n", GitCommit)
	fmt.Printf("Git branch: %s\n", GitBranch)
	fmt.Printf("Git tag: %s\n", GitTag)
	fmt.Printf("Git tree state: %s\n", GitDirty)
	fmt.Printf("Built %s\n", BuildTime)
}
