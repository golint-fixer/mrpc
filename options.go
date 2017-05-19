package mrpc

import "net/http"

// WithName is a functional option to set the service name
func WithName(name string) func(*Service) error {
	return func(s *Service) error {
		s.name = name
		return nil
	}
}

// WithGroup is a functional option to set the service group
func WithGroup(group string) func(*Service) error {
	return func(s *Service) error {
		s.group = group
		return nil
	}
}

// WithVersion is a functional option to set the service ver
func WithVersion(ver string) func(*Service) error {
	return func(s *Service) error {
		s.version = ver
		return nil
	}
}

// WithNGV is a functional option to set the service name, groupe and version
func WithNGV(name, group, ver string) func(*Service) error {
	return func(s *Service) error {
		s.name, s.group, s.version = name, group, ver
		return nil
	}
}

// WithAdapter is a functional option to set the service message adapter
func WithAdapter(a MessageAdapter) func(*Service) error {
	return func(s *Service) error {
		s.adapter = a
		return nil
	}
}

// WithStatus is a functional option to enable status endpoint
func WithStatus(addr string) func(*Service) error {
	return func(s *Service) error {
		s.statusSrv = &http.Server{Addr: addr, Handler: s}
		return nil
	}
}
