package main

import (
	"github.com/6degreeshealth/autotest"
)

func main() {
	autotest.Watch(".", autotest.RunTests, autotest.PrintTest)
}
