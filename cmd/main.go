package main

import (
	"time"

	"github.com/6degreeshealth/autotest"
)

func main() {
	autotest.Watch(250*time.Millisecond, ".", RunTests("."))
}
