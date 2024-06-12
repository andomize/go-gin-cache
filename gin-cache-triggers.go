package gogincache

import (
	"net/http"
	"regexp"
)

type Trigger interface {
	Comparable(*http.Request) bool
}

var (
	DefaultUpdateMethods = []string{
		http.MethodPost,
		http.MethodPatch,
		http.MethodPut,
	}
)

type TriggerURI struct {
	Methods []string // HTTP methods
	URI     string   // Regular expression
}

func (tURI *TriggerURI) Comparable(r *http.Request) bool {
	// Сравниваем URI по частичному совпадению.
	re := regexp.MustCompile(tURI.URI)
	match := re.MatchString(r.RequestURI)
	if match {
		for _, method := range tURI.Methods {
			if method == r.Method {
				return true
			}
		}
	}

	return false
}
