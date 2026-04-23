package main

import (
	"fmt"
	"os"

	"github.com/bssm-oss/web-replica/internal/cli"
)

func main() {
	if err := cli.NewWebReplicaCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
