package stackb

import "net/http"

func ListenAndServe(addr, version string) error {
	s := New(version)
	return http.ListenAndServe(addr, s.Handler())
}
