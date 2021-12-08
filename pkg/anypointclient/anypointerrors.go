package anypointclient

import "fmt"

/*
Err represent a error from the AnypointClient library
*/
type Err struct {
	message string
}

/*
Errorf will create a new AnypointClientError
*/
func Errorf(format string, args ...interface{}) *Err {
	return &Err{
		message:fmt.Sprintf(format, args...),
	}
}

func (e *Err) Error() string {
    return e.message
}
