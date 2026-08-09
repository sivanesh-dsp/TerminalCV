package shell

import (
	_ "embed"
	"strings"
)

//go:embed banner.txt
var bannerLarge string

//go:embed banner-small.txt
var bannerSmall string

// bannerFor returns an ASCII wordmark sized to the terminal width.
func bannerFor(width int) string {
	switch {
	case width >= 74:
		return strings.TrimRight(bannerLarge, "\n")
	case width >= 46:
		return strings.TrimRight(bannerSmall, "\n")
	default:
		return "SIVANESH B"
	}
}
