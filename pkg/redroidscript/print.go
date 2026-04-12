package redroidscript

import "fmt"

const (
	green  = "\033[32m"
	yellow = "\033[33m"
	endc   = "\033[0m"
)

func printGreen(s string) { fmt.Println(green + s + endc) }
func printYellow(s string) { fmt.Println(yellow + s + endc) }
