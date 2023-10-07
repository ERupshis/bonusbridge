package validator

import (
	"strconv"
)

// IsLuhnValid checks if a given number is valid according to the Luhn algorithm.
func IsLuhnValid(number string) bool {
	var sum int
	// Determine if we need to double every other digit or not
	double := false

	// Iterate through the digits in reverse order
	for i := len(number) - 1; i >= 0; i-- {
		digit, err := strconv.Atoi(string(number[i]))
		if err != nil {
			// Invalid digit
			return false
		}

		// Double every other digit
		if double {
			digit *= 2
			if digit > 9 {
				// Subtract 9 from numbers greater than 9
				digit -= 9
			}
		}

		sum += digit
		double = !double
	}

	// Check if the sum is a multiple of 10
	return sum%10 == 0
}
