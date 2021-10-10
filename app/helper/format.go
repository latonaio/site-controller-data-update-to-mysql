package helper

import "strings"

// PostalCodeFormat format the zip code; 1234567 to 123-4567
func PostalCodeFormat(postalCode string) string {
	if strings.Contains(postalCode, "-") {
		return postalCode
	}
	return postalCode[:3] + "-" + postalCode[3:]
}
