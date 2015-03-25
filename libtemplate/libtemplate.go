package libtemplate

import (
	"github.com/GeertJohan/go.rice"
	"html/template"
	"strings"
)

var CachedTemplates = make(map[string]*template.Template)

func GetFromRicebox(box *rice.Box, useCache bool, relativePaths ...string) (*template.Template, error) {
	templateKey := strings.Join(relativePaths, ",")

	// Check cache first and return if result is found
	if useCache {
		fromCache, ok := CachedTemplates[templateKey]
		if ok && fromCache != nil {
			return fromCache, nil
		}
	}

	tmpl := template.New(templateKey)

	for _, relativePath := range relativePaths {

		templateString, err := box.String(relativePath)
		if err != nil {
			return nil, err
		}

		tmpl, err = tmpl.Parse(templateString)
		if err != nil {
			return nil, err
		}
	}

	CachedTemplates[templateKey] = tmpl

	return tmpl, nil
}
