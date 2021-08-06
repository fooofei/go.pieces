package sshttp

import (
	"os"
)

// ssh -o "ProxyCommand " 是通过 stdin/stdout 来传递数据

type SSHPipeRW struct {
}

func (sp *SSHPipeRW) Read(p []byte) (int, error) {
	return os.Stdin.Read(p)
}
func (sp *SSHPipeRW) Write(p []byte) (int, error) {
	return os.Stdout.Write(p)
}
