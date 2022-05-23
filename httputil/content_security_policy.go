package httputil

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

type hostRequiredSpec struct {
	addThemePark   bool
	addGitHub      bool
	addFontAwesome bool
	noneRequired   bool
}

const (
	themeParkHost   = "theme-park.dev"
	rawGitHubHost   = "raw.githubusercontent.com"
	fontAwesomeHost = "use.fontawesome.com"
	styleSrc        = "style-src"
	imgSrc          = "img-src"
	fontSrc         = "font-src"
	formAction      = "form-action"
	frameAncestors  = "frame-ancestors"

	contentSecurityPolicy = "Content-Security-Policy"

	defaultFormAction     = "form-action 'self'"
	defaultFrameAncestors = "frame-ancestors 'self'"
	defaultFontAwesome    = "font-src use.fontawesome.com"
)

// EnsureProperContentSecurityPolicy ensure that if the http.Header contains
// "Content-Security-Policy" it also contains required references for support.
func EnsureProperContentSecurityPolicy(header *http.Header) {
	policy := header.Get(contentSecurityPolicy)
	if policy == "" {
		return
	}

	log.Printf("CSP: %v", policy)

	categories := strings.Split(policy, "; ")

	for index, item := range categories {
		fullSplit := strings.Split(item, " ")
		policyName := fullSplit[0]
		policyContent := fullSplit[1:]

		requiredHosts := getHostRequiredSpecForPolicy(policyName)

		if !requiredHosts.noneRequired {
			categories[index] = addMissingHosts(policyName, policyContent, requiredHosts)
		}
	}

	properHeader := strings.Join(categories, "; ")

	header.Set(contentSecurityPolicy, properHeader)
}

func addMissingHosts(policyName string, policyContent []string, requiredHosts hostRequiredSpec) string {

	for _, item := range policyContent {
		switch item {
		case themeParkHost:
			requiredHosts.addThemePark = false
		case rawGitHubHost:
			requiredHosts.addGitHub = false
		case fontAwesomeHost:
			requiredHosts.addFontAwesome = false
		}
	}

	if requiredHosts.addThemePark {
		policyContent = append(policyContent, themeParkHost)
	}

	if requiredHosts.addGitHub {
		policyContent = append(policyContent, rawGitHubHost)
	}

	if requiredHosts.addFontAwesome {
		policyContent = append(policyContent, fontAwesomeHost)
	}
	log.Printf("'%s' %v", policyName, policyContent)

	return fmt.Sprintf("%s %s", policyName, strings.Join(policyContent, " "))
}

func getHostRequiredSpecForPolicy(policyName string) hostRequiredSpec {
	switch policyName {
	case styleSrc:
		return hostRequiredSpec{
			addThemePark:   true,
			addGitHub:      true,
			addFontAwesome: true,
		}

	case imgSrc:
		return hostRequiredSpec{
			addThemePark: true,
			addGitHub:    true,
		}

	case fontSrc:
		return hostRequiredSpec{
			addFontAwesome: true,
		}

	default:
		// If the current policy has no required elements carry on as is
		return hostRequiredSpec{
			noneRequired: true,
		}
	}
}

func requiredPolicyNames() []string {
	return []string{
		"default-src",
		"style-src",
		"img-src",
		"script-src",
		"object-src",
		"form-action",
		"frame-ancestors",
		"font-src",
	}
}
