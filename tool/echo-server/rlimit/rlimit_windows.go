// +build windows

package rlimit

// BreakOpenFilesLimit will act an empty func on Windows.
func BreakOpenFilesLimit() {
	// log.Printf("called Windows version of BreakOpenFilesLimit")
}
