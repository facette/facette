package utils

import (
	"flag"
)

var (
	flagConfig string
)

func init() {
	flag.StringVar(&flagConfig, "c", "", "configuration file path")
}
