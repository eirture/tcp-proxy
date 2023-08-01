package build

import (
	"fmt"
	"time"
)

var (
	Version string
	Date    string = time.Now().Format("2006-01-02")
)

func PrintVersion() {
	fmt.Printf("Version:%s\n", VersionWithDate())
}

func VersionWithDate() string {
	return fmt.Sprintf("%s - (%s)", Version, Date)
}
