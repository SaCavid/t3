package main

import (
	"t3/api"
)

func init() {
	api.DescribeTitle()
	api.Startup()
}

func main() {
	api.Listen()
}
