package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/nyaruka/phonenumbers"
	"github.com/spf13/cobra"
	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
)

// getCountryName converts an ISO region code (e.g., "DE") to its full English country name (e.g., "Germany").
// It uses a direct lookup for the region name.
func getCountryName(regionCode string) string {
	if regionCode == "" {
		return "Unknown Region"
	}

	// Directly parse the regionCode into a language.Region object.
	// This is preferred for looking up region-specific names.
	parsedRegion, err := language.ParseRegion(regionCode)
	if err != nil {
		// If parsing the region code fails (unusual for valid codes from phonenumbers),
		// return the original code as a fallback.
		return regionCode
	}

	// Get the English display name for the parsed region.
	name := display.English.Regions().Name(parsedRegion)

	// If the name is empty, it might mean display data is unavailable for this region code.
	// For "ZZ" (Unknown/Invalid), explicitly return "Unknown Region".
	// For other codes, return the code itself as a fallback.
	if name == "" {
		if regionCode == "ZZ" {
			return "Unknown Region"
		}
		return regionCode
	}

	return name
}

// luPnCmd represents the luPn command
var luPnCmd = &cobra.Command{
	Use:   "luPn [phone number or prefix]", // Corrected spacing
	Short: "Look up country info from a phone number or prefix",
	Long: `luPn (Lookup Phone Number) works with complete numbers like +4912345678
or country prefixes like +822 to return country information.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		input := args[0]

		// Ensure input starts with "+" for international format parsing.
		if !strings.HasPrefix(input, "+") {
			input = "+" + input
		}

		// Attempt to parse the input as a full phone number.
		num, err := phonenumbers.Parse(input, "") // Default region "" means expect international format.
		if err == nil {
			// Successfully parsed as a full number.
			region := phonenumbers.GetRegionCodeForNumber(num)
			countryName := getCountryName(region)
			isValid := phonenumbers.IsValidNumber(num)
			isPossible := phonenumbers.IsPossibleNumber(num)

			fmt.Println("📞 Phone Number Analysis:")
			fmt.Printf("• Input: %s\n", input)
			fmt.Printf("• Country Code: +%d\n", num.GetCountryCode())
			fmt.Printf("• Country: %s\n", countryName)
			fmt.Printf("• National Number: %d\n", num.GetNationalNumber())
			fmt.Printf("• Valid: %t\n", isValid)
			fmt.Printf("• Possibly Valid: %t\n", isPossible)
			return
		}

		// Fallback: If full parse fails, try to identify by country code prefix.
		// This handles inputs like "+49" or "44".
		rawInput := strings.TrimPrefix(input, "+")
		if rawInput == "" { // Handle case where input was just "+" or became empty.
			fmt.Println("❌ Could not identify country or region for input:", input)
			fmt.Println("Hint: Try a valid prefix like +49 or a full number like +4912345678")
			return
		}

		var identifiedPrefixInfo bool
		// Iterate through possible prefix lengths (max 4 digits, common for country codes, down to 1).
		for i := min(len(rawInput), 4); i >= 1; i-- {
			codeCandidateStr := rawInput[:i]
			codeCandidateInt, convErr := strconv.Atoi(codeCandidateStr)
			if convErr != nil {
				// Current candidate substring is not a number; try shorter.
				continue
			}

			regions := phonenumbers.GetRegionCodesForCountryCode(codeCandidateInt)
			if len(regions) > 0 {
				var countryNames []string
				for _, region := range regions {
					countryNames = append(countryNames, getCountryName(region))
				}
				fmt.Println("📞 Partial Match (Country Code only):")
				fmt.Printf("• Input: %s\n", input)
				fmt.Printf("• Identified Country Code: +%d\n", codeCandidateInt)
				fmt.Printf("• Possible Countries/Regions: %s\n", strings.Join(countryNames, ", "))
				identifiedPrefixInfo = true
				break // Found the longest valid prefix; no need to check shorter ones.
			}
		}

		if !identifiedPrefixInfo {
			fmt.Println("❌ Could not identify country or region for input:", input)
			fmt.Println("Hint: Try a valid prefix like +49 or a full number like +4912345678")
		}
	},
}

// min is a helper function to find the minimum of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func init() {
	// Assumes rootCmd is defined in another file in the same package (e.g., cmd/root.go)
	rootCmd.AddCommand(luPnCmd)
}
