package errs

import (
	"errors"
	"fmt"
)

type Factory interface {
	Message(fstr string, args ...any) Factory
	UserMessage(fstr string, args ...any) Factory
	Logs([]any) Factory
	Mark(...error) Factory
	Private() Factory
	Public() Factory
	Domain(string) Factory
	Err() error
}

type factory struct {
	internal    error
	safeMessage string
	logDetails  []any
	userDetails any
	domain      string
	markers     []error

	private bool  // effective
	forced  *bool // nil = auto, non-nil = locked
}

func F() Factory {
	return &factory{
		private:    true, // default
		logDetails: make([]any, 0),
		markers:    make([]error, 0),
	}
}

func (f *factory) clone() *factory {
	cp := *f
	if f.logDetails != nil {
		cp.logDetails = append([]any{}, f.logDetails...)
	}
	if f.markers != nil {
		cp.markers = append([]error{}, f.markers...)
	}
	return &cp
}

func (f *factory) Message(fstr string, args ...any) Factory {
	cp := f.clone()
	cp.internal = fmt.Errorf(fstr, args...)
	return cp
}

func (f *factory) UserMessage(fstr string, args ...any) Factory {
	cp := f.clone()
	cp.safeMessage = fmt.Sprintf(fstr, args...)
	return cp
}

func (f *factory) Logs(v []any) Factory {
	cp := f.clone()
	cp.logDetails = append(cp.logDetails, v...)
	return cp
}

func (f *factory) Mark(errs ...error) Factory {
	cp := f.clone()
	cp.markers = append(cp.markers, errs...)

	// If visibility was explicitly forced, do not infer.
	if cp.forced != nil {
		return cp
	}

	for _, e := range errs {
		code := GetHTTPCode(e)
		if code < 500 {
			cp.private = false
		}
	}

	return cp
}

func (f *factory) Private() Factory {
	cp := f.clone()
	v := true
	cp.private = true
	cp.forced = &v
	return cp
}

func (f *factory) Public() Factory {
	cp := f.clone()
	v := false
	cp.private = false
	cp.forced = &v
	return cp
}

func (f *factory) Domain(d string) Factory {
	cp := f.clone()
	cp.domain = d
	return cp
}

func (f *factory) Err() error {
	if f.internal == nil {
		f.internal = errors.New("unknown error")
	}

	return &Error{
		Internal:       f.internal,
		ExposeInternal: !f.private,
		SafeMessage:    f.safeMessage,
		LogDetails:     f.logDetails,
		UserDetails:    f.userDetails,
		Domain:         f.domain,
		Markers:        f.markers,
	}
}
