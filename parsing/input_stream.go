package parsing

type InputStream struct {
	input []rune
	pos int
	line int
	col int
}

func NewInputStream(input string) *InputStream {
	return &InputStream{
		input: []rune(input),
		pos: 0,
		line: 1,
		col: 0,
	}
}

func (is *InputStream) next() rune {
	ch := is.input[is.pos]
	is.pos++
	if ch == '\n' {
		is.line++
		is.col = 0
	} else {
		is.col++
	}
	return ch
}

func (is *InputStream) peek() rune {
	return is.input[is.pos]
}

func (is *InputStream) eof() bool {
	return is.pos >= len(is.input)
}
