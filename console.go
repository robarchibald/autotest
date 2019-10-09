package autotest

import (
	"bufio"
	"bytes"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/logrusorgru/aurora"
)

// PrintTest is used to print a single test results to the console
func PrintTest(result *TestResult) {
	margin := (80 - len(result.Folder)) / 2
	fmt.Println()
	fmt.Println(strings.Repeat("-", margin), result.Folder, strings.Repeat("-", margin))
	if result.BuildFail != nil {
		printBuildFailure(result.BuildFail)
	}
	if len(result.Status) != 0 {
		printTestEvents(result.Status, result.TestFail)
	}
	if len(result.Coverage) != 0 {
		printCoverage(result.Coverage)
	}
}

func printTestEvents(groupedEvents []TestStatus, showAll bool) {
	groupedEvents, maxPackageLen, maxTestLen := getFilteredListAndLengths(groupedEvents, showAll)
	if len(groupedEvents) != 0 {
		printHeader("--- Test Results ---", "Time  ", rightPad("Package", maxPackageLen), rightPad("Test", maxTestLen), "Status")
	}
	for _, event := range groupedEvents {
		fmt.Println(printElapsedTime(event.Elapsed), rightPad(getPackage(event.Package), maxPackageLen), aurora.BrightWhite(rightPad(getTestName(event.Test), maxTestLen)), printTestResult(event.TestResult), printOutput(event.Output))
	}
}

func printHeader(header string, columns ...string) {
	totalWidth := 0
	for _, column := range columns {
		totalWidth += len(column) + 1
	}
	fmt.Println(strings.Repeat(" ", (totalWidth-len(header))/2), aurora.Blue(header))
	for _, column := range columns {
		fmt.Print(aurora.Gray(15, column+" "))
	}
	fmt.Println()
}

func getFilteredListAndLengths(groupedEvents []TestStatus, showAll bool) ([]TestStatus, int, int) {
	maxPackageLen := 0
	maxTestLen := len("[package]")
	filteredList := []TestStatus{}
	for _, event := range groupedEvents {
		if event.Test == "" || showAll || event.Elapsed > 0.1 {
			filteredList = append(filteredList, event)
			if l := len(getPackage(event.Package)); l > maxPackageLen {
				maxPackageLen = l + 1
			}
			if l := len(event.Test); l > maxTestLen {
				maxTestLen = l + 1
			}
		}
	}
	return filteredList, maxPackageLen, maxTestLen
}

var buildFailParse = regexp.MustCompile(`(^.*?):(\d*):(\d*):(.*)$`) // <file info><line number>:<column number>:<error message>

func printBuildFailure(out []byte) {
	r := bufio.NewScanner(bytes.NewReader(out))
	for r.Scan() {
		line := r.Text()
		if hasAnyPrefix(line, []string{"#", "FAIL"}) {
			continue
		}
		parsedLine := buildFailParse.FindStringSubmatch(line)
		if len(parsedLine) > 0 {
			pathInfo := parsedLine[1]
			lastColon := strings.LastIndex(pathInfo, ":")
			if lastColon != -1 {
				pathInfo = strings.TrimSpace(pathInfo[lastColon+1:])
			}
			fmt.Printf("Error in %s at line %s, column %s\n%s\n", pathInfo, aurora.Blue(parsedLine[2]), aurora.Blue(parsedLine[3]), aurora.Red(strings.TrimSpace(parsedLine[4])))
		}
	}
}

func printElapsedTime(elapsed float64) string {
	color := aurora.WhiteFg
	if elapsed > 0.5 {
		color = aurora.RedFg
	} else if elapsed > 0.1 {
		color = aurora.YellowFg
	}
	return aurora.Colorize(formatFloat(elapsed, 2)+"s ", color).String()
}

func formatFloat(num float64, decimalPlaces int) string {
	return strconv.FormatFloat(math.Round(num*math.Pow(10.0, float64(decimalPlaces)))/math.Pow(10.0, float64(decimalPlaces)), 'f', decimalPlaces, 64)
}

func getPackage(pkg string) string {
	split := strings.Split(pkg, "/")
	if len(split) < 3 {
		return pkg
	}
	return strings.Join(split[2:], "/")
}

func rightPad(text string, length int) string {
	return fmt.Sprintf(fmt.Sprintf("%%-%ds", length), text)
}

func getTestName(test string) string {
	if test == "" {
		return "[package]"
	}
	return test
}

func printTestResult(status string) string {
	switch status {
	case "pass":
		return aurora.Green("PASS").String()
	case "fail":
		return aurora.Red("FAIL").String()
	case "skipped":
		return aurora.Yellow("SKIPPED").String()
	}
	return status
}

func printOutput(output string) string {
	if len(output) > 0 {
		return aurora.Gray(10, fmt.Sprintf("\noutput:%s\n", output)).String()
	}
	return ""
}

func printCoverage(coverageItems []FunctionCoverage) {
	maxFilenameLen, maxFunctionLen, not100Percent := getCoverageLengths(coverageItems)
	if not100Percent == 0 {
		return
	}
	printHeader("--- Code Coverage ---", rightPad("Filename", maxFilenameLen), rightPad("Function", maxFunctionLen), "Coverage")
	for _, coverage := range coverageItems {
		if coverage.CoveragePercent == 100 {
			continue
		}
		fmt.Println(rightPad(coverage.Filename, maxFilenameLen), rightPad(coverage.Function, maxFunctionLen), printPercent(float64(coverage.CoveragePercent)))
	}
}

func getCoverageLengths(coverageItems []FunctionCoverage) (int, int, int) {
	var maxFilenameLen, maxFunctionLen, not100Percent int
	for _, coverage := range coverageItems {
		if l := len(coverage.Filename); l > maxFilenameLen {
			maxFilenameLen = l
		}
		if l := len(coverage.Function); l > maxFunctionLen {
			maxFunctionLen = l
		}
		if coverage.CoveragePercent != 100 {
			not100Percent++
		}
	}
	return maxFilenameLen, maxFunctionLen, not100Percent
}

func printPercent(percent float64) string {
	if percent == 100 {
		return aurora.Green("100%").String()
	} else if percent > 75 {
		return aurora.Yellow(formatFloat(percent, 1) + "%").String()
	}
	return aurora.BrightRed(formatFloat(percent, 1) + "%").String()
}
