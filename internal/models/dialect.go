package models

import (
	"strconv"
	"strings"
)

func bindQuery(query, dialect string) string {
	if dialect != "postgres" {
		return query
	}

	var bound strings.Builder
	bound.Grow(len(query) + 8)
	parameter := 1
	for _, character := range query {
		if character == '?' {
			bound.WriteByte('$')
			bound.WriteString(strconv.Itoa(parameter))
			parameter++
			continue
		}
		bound.WriteRune(character)
	}
	return bound.String()
}
