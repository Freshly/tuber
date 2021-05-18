package db

import (
	"strconv"
)

func Q() Query {
	return Query{}
}

type Query struct {
	Strings map[string]string
	Bools   map[string]bool
	Ints    map[string]int
}

func (q Query) String(k string, v string) Query {
	q.Strings[k] = v
	return q
}

func (q Query) Bool(k string, v bool) Query {
	q.Bools[k] = v
	return q
}

func (q Query) Int(k string, v int) Query {
	q.Ints[k] = v
	return q
}

func (q Query) normalize() map[string]string {
	var convertedQueryVals map[string]string

	for k, v := range q.Strings {
		convertedQueryVals[k] = v
	}

	for k, v := range q.Bools {
		convertedQueryVals[k] = strconv.FormatBool(v)
	}

	for k, v := range q.Ints {
		convertedQueryVals[k] = strconv.Itoa(v)
	}

	return convertedQueryVals
}
