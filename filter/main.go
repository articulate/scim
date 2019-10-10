package filter

import (
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	filter "github.com/di-wu/scim-filter-parser"
)

const (
	placeholder = "?"
)

var comparatorMap = map[filter.Token]string{
	filter.EQ:  "=",
	filter.NE:  "!=",
	filter.CO:  "LIKE",
	filter.SW:  "LIKE",
	filter.EW:  "LIKE",
	filter.PR:  "IS NOT NULL",
	filter.GT:  ">",
	filter.GE:  ">=",
	filter.LT:  "<",
	filter.LE:  "<=",
	filter.AND: "AND",
	filter.OR:  "OR",
	filter.NOT: "NOT",
}

var likeExpressionMap = map[filter.Token]likeExpression{
	filter.CO: likeExpression{Prefix: "%", Suffix: "%"},
	filter.SW: likeExpression{Prefix: "", Suffix: "%"},
	filter.EW: likeExpression{Prefix: "%", Suffix: ""},
}

type (
	likeExpression struct {
		Prefix string
		Suffix string
	}

	Parser struct {
		Expression filter.Expression
		RawFilter  string
		params     []string
	}
)

// NewParser builds a new filter parser with the ability to transpile filter params to SQL queries.
func NewParser(rawFilter string) (*Parser, error) {
	var exp filter.Expression

	rawFilter = strings.TrimSpace(rawFilter)

	if rawFilter != "" {
		var err error

		parser := filter.NewParser(strings.NewReader(rawFilter))
		exp, err = parser.Parse()

		if err != nil {
			return nil, err
		}
	}

	return &Parser{
		Expression: exp,
		RawFilter:  rawFilter,
	}, nil
}

// ToSql transpiles parsed filter to a SQL query. The attribute map is used to map
// schema properties to database columns.
// Example:
// sqlBuilder := parser.ToSql(
//	  map[string]string{"userName": "users.username"},
//    "users",
//    []string{"LEFT JOIN emails ON emails.user_id = users.id"},
// )
func (p *Parser) ToSql(attributeMap map[string]string, tableName string, joins []string) sq.SelectBuilder {
	baseQuery := sq.Select("*").From(tableName)

	for _, join := range joins {
		baseQuery = baseQuery.JoinClause(join)
	}

	whereClause := p.process(p.Expression, attributeMap)

	return baseQuery.Where(whereClause, p.params)
}

func (p *Parser) process(exp filter.Expression, attrMap map[string]string) string {
	if attrExp, ok := exp.(filter.AttributeExpression); ok {
		return p.processAttributeStatement(attrExp, attrMap)
	}

	if biExp, ok := exp.(filter.BinaryExpression); ok {
		return fmt.Sprintf(
			"(%s %s %s)",
			p.process(biExp.X, attrMap),
			getComparator(biExp.CompareOperator),
			p.process(biExp.Y, attrMap),
		)
	}

	if uniExp, ok := exp.(filter.UnaryExpression); ok {
		return fmt.Sprintf("(%s %s)", getComparator(uniExp.CompareOperator), p.process(uniExp.X, attrMap))
	}

	// Should never happen but handled nonetheless
	panic("unsupported expression type")
}

func (p *Parser) processAttributeStatement(exp filter.AttributeExpression, attrMap map[string]string) string {
	path := p.processAttributePath(exp.AttributePath, attrMap)
	comparator := getComparator(exp.CompareOperator)
	value := p.processAttributeValue(exp.CompareValue, exp.CompareOperator)

	return fmt.Sprintf("%s %s %s", path, comparator, value)
}

func (p *Parser) processAttributePath(path string, attrMap map[string]string) string {
	return attrMap[path]
}

func (p *Parser) processAttributeValue(value string, op filter.Token) string {
	p.params = append(p.params, value)
	tokens, ok := likeExpressionMap[op]

	if !ok {
		tokens = likeExpression{}
	}

	return tokens.Prefix + placeholder + tokens.Suffix
}

func getComparator(comparator filter.Token) string {
	if v, ok := comparatorMap[comparator]; ok {
		return v
	}

	// Should never happen but handled nonetheless
	panic("unsupported expression comparator")
}
