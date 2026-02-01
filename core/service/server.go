package service

import (
	"fmt"
	"sync"
)

// Server represents a single backend server instance
type Server struct {
	// Required fields
	Name string `config:"name"`
	Port int64  `config:"port"`

	// Optional fields
	Protocol string `config:"protocol"`
	Weight   int64  `config:"weight,default=1"`
	Disabled bool   `config:"disabled,default=false"`

	// Optional override configurations
	Request     *Request     `config:"request"`
	Response    *Response    `config:"response"`
	Auth        *Auth        `config:"auth"`
	HealthCheck *HealthCheck `config:"health_check"`

	// Runtime state (not serialized)
	healthy     bool
	healthMutex sync.RWMutex
}

// IsHealthy returns the health status of the server
func (s *Server) IsHealthy() bool {
	s.healthMutex.RLock()
	defer s.healthMutex.RUnlock()
	return s.healthy
}

// SetHealthy sets the health status of the server
func (s *Server) SetHealthy(healthy bool) {
	s.healthMutex.Lock()
	defer s.healthMutex.Unlock()
	s.healthy = healthy
}

// ID returns a unique identifier for the server
func (s *Server) ID() string {
	return fmt.Sprintf("%s:%d", s.Name, s.Port)
}

// Host returns the host:port string
func (s *Server) Host() string {
	port := s.Port
	if port == 0 {
		port = 80
	}
	return fmt.Sprintf("%s:%d", s.Name, port)
}

// Target returns the full target URL
func (s *Server) Target() string {
	protocol := s.Protocol
	if protocol == "" {
		protocol = "http"
	}
	return fmt.Sprintf("%s://%s", protocol, s.Host())
}

// GetEffectiveConfig merges the server's override configurations with the base service configuration
func (s *Server) GetEffectiveConfig(base *Service) *Service {
	effective := *base

	// Override protocol if specified
	if s.Protocol != "" {
		effective.Protocol = s.Protocol
	}

	// Merge request configuration
	if s.Request != nil {
		effective.Request = mergeRequest(base.Request, *s.Request)
	}

	// Merge response configuration
	if s.Response != nil {
		effective.Response = mergeResponse(base.Response, *s.Response)
	}

	// Override auth configuration
	if s.Auth != nil {
		effective.Auth = *s.Auth
	}

	// Merge health check configuration
	if s.HealthCheck != nil {
		effective.HealthCheck = mergeHealthCheck(base.HealthCheck, *s.HealthCheck)
	}

	return &effective
}

// mergeRequest merges two Request configurations
func mergeRequest(base, override Request) Request {
	merged := base

	// Merge path configuration
	if override.Path.Rewrites != nil {
		merged.Path.Rewrites = override.Path.Rewrites
	}
	if override.Path.DisablePrefixRewrite {
		merged.Path.DisablePrefixRewrite = override.Path.DisablePrefixRewrite
	}

	// Merge headers
	if override.Headers != nil {
		if merged.Headers == nil {
			merged.Headers = make(map[string]string)
		}
		for k, v := range override.Headers {
			merged.Headers[k] = v
		}
	}

	// Merge query parameters
	if override.Query != nil {
		if merged.Query == nil {
			merged.Query = make(map[string]string)
		}
		for k, v := range override.Query {
			merged.Query[k] = v
		}
	}

	return merged
}

// mergeResponse merges two Response configurations
func mergeResponse(base, override Response) Response {
	merged := base

	// Merge headers
	if override.Headers != nil {
		if merged.Headers == nil {
			merged.Headers = make(map[string]string)
		}
		for k, v := range override.Headers {
			merged.Headers[k] = v
		}
	}

	return merged
}

// mergeHealthCheck merges two HealthCheck configurations
func mergeHealthCheck(base, override HealthCheck) HealthCheck {
	merged := base

	if override.Enable {
		merged.Enable = override.Enable
	}
	if override.Method != "" {
		merged.Method = override.Method
	}
	if override.Path != "" {
		merged.Path = override.Path
	}
	if override.Status != nil {
		merged.Status = override.Status
	}
	if override.Interval > 0 {
		merged.Interval = override.Interval
	}
	if override.Timeout > 0 {
		merged.Timeout = override.Timeout
	}
	if override.Ok {
		merged.Ok = override.Ok
	}

	return merged
}

// IsMultiServer checks if the service is configured for multi-server mode
func (s *Service) IsMultiServer() bool {
	return len(s.Servers) > 0
}
