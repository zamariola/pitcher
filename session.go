package pitcher

import "os"

type Session interface {
	Get(key string) (string, bool)
	Put(key, val string)
}

type DynamicRWSession struct {
	parameters map[string]string
}

func NewMemoryRWSession(parameters map[string]string) *DynamicRWSession {
	return &DynamicRWSession{
		parameters: parameters,
	}
}

func (de *DynamicRWSession) Get(key string) (string, bool) {

	v, ok := de.parameters[key]

	if ok {
		return v, true
	}

	osEnv := os.Getenv(key)
	return osEnv, len(osEnv) > 0
}

func (de *DynamicRWSession) Put(key, value string) {
	de.parameters[key] = value
}
