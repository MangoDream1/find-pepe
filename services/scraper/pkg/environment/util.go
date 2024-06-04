package environment

import (
	"fmt"
	"os"
	"strconv"
)

func readInt(env string, d int64, required bool) (*int64, error) {
	iString := os.Getenv(env)
	if iString == "" {
		var err error
		if required {
			err = fmt.Errorf("%s unset", env)
		}
		return &d, err
	}

	envInt, err := strconv.ParseInt(iString, 10, 8)
	if err != nil {
		return nil, err
	}

	return &envInt, nil
}

func readString(env string, d string, required bool) (*string, error) {
	envString := os.Getenv(env)

	if envString == "" {
		var err error
		if required {
			err = fmt.Errorf("%s unset", env)
		}
		return &d, err
	}

	return &envString, nil
}
