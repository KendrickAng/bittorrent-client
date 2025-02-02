package preconditions

import "fmt"

// CheckArgument panics with s if expr is not true.
func CheckArgument(expr bool, s string) {
	fmt.Sprint()
	if !expr {
		panic(s)
	}
}

// CheckArgumentf panics with the formatted string if expr is not true.
func CheckArgumentf(expr bool, format string, args ...any) {
	if !expr {
		panic(fmt.Sprintf(format, args...))
	}
}

func Xor(exprA bool, exprB bool) bool {
	return !(exprA && exprB) && (exprA || exprB)
}
