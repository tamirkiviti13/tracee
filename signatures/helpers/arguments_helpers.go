package helpers

import (
	b64 "encoding/base64"
	"fmt"

	"github.com/aquasecurity/tracee/types/trace"
)

// GetTraceeArgumentByName fetches the argument in event with `Name` that matches argName
func GetTraceeArgumentByName(event trace.Event, argName string) (trace.Argument, error) {
	for _, arg := range event.Args {
		if arg.Name == argName {
			return arg, nil
		}
	}
	return trace.Argument{}, fmt.Errorf("argument %s not found", argName)
}

// GetTraceeStringArgumentByName gets the argument matching the "argName" given from the event "argv" field, casted as string.
func GetTraceeStringArgumentByName(event trace.Event, argName string) (string, error) {
	arg, err := GetTraceeArgumentByName(event, argName)
	if err != nil {
		return "", err
	}
	argStr, ok := arg.Value.(string)
	if ok {
		return argStr, nil
	}

	return "", fmt.Errorf("can't convert argument %v to string", argName)
}

// GetTraceeBytesSliceArgumentByName gets the argument matching the "argName" given from the event "argv" field, casted as []byte.
func GetTraceeBytesSliceArgumentByName(event trace.Event, argName string) ([]byte, error) {
	arg, err := GetTraceeArgumentByName(event, argName)
	if err != nil {
		return nil, err
	}
	argBytes, ok := arg.Value.([]byte)
	if ok {
		return argBytes, nil
	}

	argBytesString, ok := arg.Value.(string)
	if ok {
		decodedBytes, err := b64.StdEncoding.DecodeString(argBytesString)
		if err != nil {
			return nil, fmt.Errorf("can't convert argument %v to []bytes", argName)
		}
		return decodedBytes, nil
	}

	return nil, fmt.Errorf("can't convert argument %v to []bytes", argName)
}
