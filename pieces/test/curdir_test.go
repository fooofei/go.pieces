package test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestCurrentDirOfRuntime0(t *testing.T) {
	// only right of where the .go source file compiled directory
	// wrong of the executable file
	//

	_, callerFile, _, _ := runtime.Caller(0)
	t.Logf("runtime.Caller(0)= \n")
	t.Logf("  file= %v\n", callerFile)
	t.Logf("  dir= %v\n", filepath.Dir(callerFile))
}

func TestOsExecutable(t *testing.T) {
	// only right of the executable file path
	// wrong of the .go file path

	// when go run xx.go
	// the output path is a temp path

	cFile, _ := os.Executable()
	t.Logf("os.Executable=\n")
	t.Logf("  file= %v\n", cFile)
	t.Logf("  dir= %v\n", filepath.Dir(cFile))
}
