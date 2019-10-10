package filter

import (
	"reflect"
	"testing"
)

var (
	attrMap = map[string]string{
		"username":     "users.username",
		"emails.value": "emails.value",
		"emails":       "emails.value",
		"emails.type":  "emails.type",
	}
	joins = []string{"LEFT JOIN emails ON emails.user_id = users.id"}
)

func TestFilaterParser(t *testing.T) {
	parser, err := NewParser(`userName eq "test"`)

	if err != nil {
		t.Errorf("Expected to create a filter parser without an error but received an error %v", err)
	}

	sql := parser.ToSql(attrMap, "users", joins)

	query, params, err := sql.ToSql()

	if query != `SELECT * FROM users LEFT JOIN emails ON emails.user_id = users.id WHERE users.username = ?` {
		t.Errorf("Malformed SQL query, %v", query)
	}

	if len(params) != 1 && params[0] != "test" {
		t.Errorf(`Malformed parameters, expected ["test"], received %v`, params)
	}
}

func TestFilterParserAdancedQuery(t *testing.T) {
	parser, err := NewParser(`emails co "example.org" and (emails.type eq "work" and emails.value co "example.org")`)

	if err != nil {
		t.Errorf("Expected to create a filter parser without an error but received an error %v", err)
	}

	sql := parser.ToSql(attrMap, "users", joins)

	query, params, err := sql.ToSql()
	expectedQuery := "SELECT * FROM users LEFT JOIN emails ON emails.user_id = users.id WHERE (emails.value LIKE %?% AND (emails.type = ? AND emails.value LIKE %?%))"

	if query != expectedQuery {
		t.Errorf("Malformed SQL query, %v", query)
	}

	if len(params) != 1 && params[0] != "test" {
		t.Errorf(`Malformed parameters, expected ["test"], received %v`, params)
	}
}

func TestFilterParser(t *testing.T) {
	var tests = []struct {
		attributeMap map[string]string
		expected     string
		filter       string
		joins        []string
		params       []interface{}
	}{
		{
			filter:       `not emails co "example.com"`,
			expected:     "SELECT * FROM users WHERE (NOT users.emails LIKE %?%)",
			params:       []interface{}{[]string{"example.com"}},
			joins:        []string{},
			attributeMap: map[string]string{"emails": "users.emails"},
		},
		{
			filter:       `userName eq "test"`,
			expected:     `SELECT * FROM users LEFT JOIN emails ON emails.user_id = users.id WHERE users.username = ?`,
			params:       []interface{}{[]string{"test"}},
			joins:        joins,
			attributeMap: attrMap,
		},
		{
			filter:       `emails co "example.org" and (emails.type eq "work" and emails.value co "example.org")`,
			expected:     "SELECT * FROM users LEFT JOIN emails ON emails.user_id = users.id WHERE (emails.value LIKE %?% AND (emails.type = ? AND emails.value LIKE %?%))",
			params:       []interface{}{[]string{"example.org", "work", "example.org"}},
			joins:        joins,
			attributeMap: attrMap,
		},
	}

	for _, test := range tests {
		parser, err := NewParser(test.filter)

		if err != nil {
			t.Errorf("Expected to create a filter parser without an error but received an error %v", err)
		}

		sql := parser.ToSql(test.attributeMap, "users", test.joins)

		query, params, err := sql.ToSql()

		if query != test.expected {
			t.Errorf("Malformed SQL query, expected:\n%s\ngot:\n%s", test.expected, query)
		}

		if len(params) != len(test.params) || !reflect.DeepEqual(params, test.params) {
			t.Errorf(`Malformed parameters, expected %v, received %v`, test.params, params)
		}
	}
}
