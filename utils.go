package mail

import "fmt"

type Args interface {
	Set(name string, value interface{})
	Get(name string) interface{}
}

type DialError struct {
	Addr string
	Err  error
}

func (e *DialError) Error() string {
	return fmt.Sprintf("dialing error %s %v", e.Addr, e.Err)
}
