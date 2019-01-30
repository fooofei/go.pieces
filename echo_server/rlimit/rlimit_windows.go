// +build windows

package rlimt

// BrkOpenFilesLimit will act an empty func on Windows.
func BrkOpenFilesLimit()  {
    // log.Printf("called Windows version of BrkOpenFilesLimit")
}