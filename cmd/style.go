package cmd

import "github.com/fatih/color"

var (
	headerStyle = color.New(color.FgCyan, color.Bold).SprintFunc()
	labelStyle  = color.New(color.FgHiWhite, color.Bold).SprintFunc()
	okStyle     = color.New(color.FgGreen, color.Bold).SprintFunc()
	warnStyle   = color.New(color.FgYellow, color.Bold).SprintFunc()
)

func styledHeader(text string) string {
	return headerStyle(text)
}

func styledLabel(text string) string {
	return labelStyle(text)
}

func styledStatus(found bool) string {
	if found {
		return okStyle("found")
	}
	return warnStyle("missing")
}

func styledMarker(found bool) string {
	if found {
		return okStyle("✓")
	}
	return warnStyle("✗")
}
