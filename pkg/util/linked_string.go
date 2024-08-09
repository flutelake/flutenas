package util

type LinkedRune struct {
	r rune
	p *LinkedRune
}

func NewLinkedRune(s string) *LinkedRune {
	var next *LinkedRune
	runes := []rune(s)
	for i := len(runes) - 1; i >= 0; i-- {
		r := &LinkedRune{
			r: runes[i],
			p: next,
		}
		next = r
	}
	return next
}

func (r *LinkedRune) Next() *LinkedRune { return r.p }

func (r *LinkedRune) String() string {
	runes := []rune{}
	for r != nil {
		runes = append(runes, r.r)
		r = r.Next()
	}
	return string(runes)
}
