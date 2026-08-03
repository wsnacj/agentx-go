// reference-host 是一个默认离线、无凭据、无网络和无文件写入的显式组合 Host。
package main

import (
	"context"
	"fmt"
	"os"
)

func run(arguments []string) error {
	value, err := parseConfig(arguments, os.Stderr)
	if err != nil {
		return err
	}
	host, err := newReferenceHost(value)
	if err != nil {
		return err
	}
	result, runErr := host.Run(context.Background())
	shutdownErr := host.Shutdown(context.Background())
	if runErr != nil {
		return runErr
	}
	if shutdownErr != nil {
		return shutdownErr
	}
	fmt.Println(result.Reply)
	return nil
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
