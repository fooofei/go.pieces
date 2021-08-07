package main

import (
	"fmt"
	"log"
	"os"
)

// 把一个可执行程序的命令行参数和环境变量打印出来

func main() {
	fw, err := os.OpenFile("d:\\1.log", os.O_RDWR|os.O_APPEND|os.O_CREATE, 0600)
	if err == nil {
		log.SetOutput(fw)
		defer fw.Close()
	}
	log.SetPrefix(fmt.Sprintf("pid %v ", os.Getpid()))

	log.Printf("args count %v", len(os.Args))
	for i, arg := range os.Args {
		log.Printf("-- %v = `%v`", i, arg)
	}

	envs := os.Environ()
	log.Printf("env count %v", len(envs))
	for i, env := range envs {
		log.Printf("-- %v `%v`", i, env)
	}
}
