package issues

import (
	"fmt"
)

type Issue struct {
	Key       string
	FirstLine int
	Line      int

	Name string
	File string

	Recommendations []string
}

func NewIssue(name, key, file string, fl, ll int, r []string) Issue {
	return Issue{
		Name:            name,
		Key:             key,
		File:            file,
		FirstLine:       fl,
		Line:            ll,
		Recommendations: r,
	}
}

func (i Issue) String() string {
	if i.Recommendations == nil && len(i.Recommendations) > 0 {
		fmt.Sprintf("%s %q found (line %d and line %d). Recommendations: %s",
			i.Name, i.Key, i.FirstLine, i.Line, i.Recommendations)
	}

	return fmt.Sprintf("%s %q found (line %d and line %d)",
		i.Name, i.Key, i.FirstLine, i.Line)
}
