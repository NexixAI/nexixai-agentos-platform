package stacka

import "net/http"

func ListenAndServe(addr, version string) error {
	s, err := New(version)
	if err != nil {
		return err
	}
	return http.ListenAndServe(addr, s.Handler())
}
