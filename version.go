package FizzBuzz

import (
	_ "embed"
)

//go:generate sh get_version.sh
//go:embed version.txt
var Version string
