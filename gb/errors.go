package main

type error interface {
	String() string
}

type serror string

func (e serror) Error() string {
	return string(e)
}

func (e serror) String() string {
	return string(e)
}

type Errors struct {}
func (es Errors) New(s string) error {
	return serror(s)
}
var errors Errors
