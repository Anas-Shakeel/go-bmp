package utils

import "fmt"

// Returns the average of all given numbers n
func Average(n ...int) int {
	// Sum all numbers
	var sum int
	for _, num := range n {
		sum += num
	}

	// Divide sum by total numbers
	return sum / len(n)
}

// Print a Colored Block in terminal
func ColoredBlock(block string, red int, green int, blue int) string {
	return fmt.Sprintf("\033[48;2;%d;%d;%dm%s\033[0m", red, green, blue, block)
}
