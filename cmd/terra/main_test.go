package main

import (
	"reflect"
	"testing"
)

func TestParseLabels(t *testing.T) {
	cases := []struct {
		name     string
		labels   []string
		expected map[string]string
	}{
		{
			name:   "simple",
			labels: []string{"foo=bar"},
			expected: map[string]string{
				"foo": "bar",
			},
		},
		{
			name:   "key only",
			labels: []string{"foo"},
			expected: map[string]string{
				"foo": "",
			},
		},
		{
			name:   "multiple",
			labels: []string{"foo=bar", "app=terra"},
			expected: map[string]string{
				"foo": "bar",
				"app": "terra",
			},
		},
		{
			name:   "complex",
			labels: []string{"foo=bar", "app=terra", "keys=foo==bar"},
			expected: map[string]string{
				"foo":  "bar",
				"app":  "terra",
				"keys": "foo==bar",
			},
		},
	}

	for _, c := range cases {
		labels, err := getLabels(c.labels)
		if err != nil {
			t.Errorf("case %s error: %s", c.name, err)
		}

		if !reflect.DeepEqual(labels, c.expected) {
			t.Errorf("case %s error; expected %s; received %s", c.name, c.expected, labels)
		}
	}
}
