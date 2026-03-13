package queryeditor

import (
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/charmbracelet/lipgloss"

	"github.com/mdjarv/db/internal/tui/theme"
)

var sqlLexer chroma.Lexer

func init() {
	sqlLexer = lexers.Get("sql")
	if sqlLexer == nil {
		sqlLexer = lexers.Fallback
	}
}

func highlightSQL(line string) string {
	iter, err := sqlLexer.Tokenise(nil, line)
	if err != nil {
		return line
	}

	var sb strings.Builder
	for _, tok := range iter.Tokens() {
		s := styleForToken(tok.Type)
		sb.WriteString(s.Render(tok.Value))
	}
	return sb.String()
}

func styleForToken(t chroma.TokenType) lipgloss.Style {
	s := theme.Current().Styles
	switch t {
	case chroma.KeywordType:
		return s.Type
	case chroma.Keyword, chroma.KeywordDeclaration, chroma.KeywordNamespace,
		chroma.KeywordReserved, chroma.KeywordConstant, chroma.KeywordPseudo:
		return s.Keyword
	case chroma.LiteralString, chroma.LiteralStringSingle,
		chroma.LiteralStringAffix, chroma.LiteralStringEscape:
		return s.String
	case chroma.LiteralNumber, chroma.LiteralNumberFloat,
		chroma.LiteralNumberInteger:
		return s.Number
	case chroma.Comment, chroma.CommentSingle, chroma.CommentMultiline:
		return s.Comment
	case chroma.NameFunction, chroma.NameBuiltin:
		return s.Function
	case chroma.Operator, chroma.OperatorWord:
		return s.Operator
	default:
		return lipgloss.NewStyle()
	}
}
