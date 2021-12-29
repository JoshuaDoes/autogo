package autoit

func (vm *AutoItVM) GetMacro(macro string) (*Token, error) {
	switch macro {
	case "CR":
		return NewToken(tSTRING, "\r"), nil
	case "LF":
		return NewToken(tSTRING, "\n"), nil
	case "CRLF":
		return NewToken(tSTRING, "\r\n"), nil
	}
	return nil, vm.Error("illegal macro: %s", macro)
}