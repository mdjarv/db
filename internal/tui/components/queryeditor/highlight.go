package queryeditor

import (
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/charmbracelet/lipgloss"
)

var (
	keywordStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
	stringStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("142"))
	numberStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
	commentStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
	typeStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("37"))
	funcStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("178"))
	opStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
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
	switch t {
	case chroma.KeywordType:
		return typeStyle
	case chroma.Keyword, chroma.KeywordDeclaration, chroma.KeywordNamespace,
		chroma.KeywordReserved, chroma.KeywordConstant, chroma.KeywordPseudo:
		return keywordStyle
	case chroma.LiteralString, chroma.LiteralStringSingle,
		chroma.LiteralStringAffix, chroma.LiteralStringEscape:
		return stringStyle
	case chroma.LiteralNumber, chroma.LiteralNumberFloat,
		chroma.LiteralNumberInteger:
		return numberStyle
	case chroma.Comment, chroma.CommentSingle, chroma.CommentMultiline:
		return commentStyle
	case chroma.NameFunction, chroma.NameBuiltin:
		return funcStyle
	case chroma.Operator, chroma.OperatorWord:
		return opStyle
	default:
		return lipgloss.NewStyle()
	}
}
