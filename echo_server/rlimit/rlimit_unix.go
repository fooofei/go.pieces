// +build !windows

package rlimt

import (
    "log"
    "syscall"
)

// BrkOpenFilesLimit will call setrlimit
func BrkOpenFilesLimit() {

    //log.Printf("called unix version of BrkOpenFilesLimit")

    var err error
    var rlim syscall.Rlimit
    var limit uint64 = 100 * 1000
    err = syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rlim)
    if err != nil {
        log.Fatalf("Getrlimit err= %v", err)
    }
    rlim.Cur = limit + uint64(100)
    rlim.Max = limit + uint64(100)
    err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rlim)
    if err != nil {
        log.Fatalf("Setrlimit err= %v", err)
    }
}
