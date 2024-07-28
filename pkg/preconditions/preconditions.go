package preconditions

import "fmt"

func CheckArgumentM(expr bool, msg string) {
	if !expr {
		panic(msg)
	}
}

func CheckArgument(expr bool) {
	if !expr {
		panic(fmt.Errorf("expected true, got false"))
	}
}
