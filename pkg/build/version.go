package build

import "fmt"

var (
	Version   string
	BuildDate string
)

func PrintVersion() {
	fmt.Printf("Version: %s - (%s)\n", Version, BuildDate)
}
