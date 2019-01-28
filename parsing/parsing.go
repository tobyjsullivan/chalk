package parsing

func Parse(formula string) (*ASTNode, error) {
	p := NewParser(NewLexer(NewInputStream(formula)))

	return p.Parse()
}
