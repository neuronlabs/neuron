package query

func (s *Scope) logFormat(format string) string {
	return "SCOPE[" + s.ID.String() + "]" + format
}
