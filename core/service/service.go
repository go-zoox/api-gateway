package service

type Service struct {
	Name     string `config:"name"`
	Port     int64  `config:"port"`
	Protocol string `config:"protocol,default=http"`
	//
	Request  Request  `config:"request"`
	Response Response `config:"response"`
	//
	Auth Auth `config:"auth"`
	//
	HealthCheck HealthCheck `config:"health_check"`
}

type Request struct {
	Path    RequestPath       `config:"path"`
	Headers map[string]string `config:"headers"`
	Query   map[string]string `config:"query"`
}

type RequestPath struct {
	DisablePrefixRewrite bool     `config:"disable_prefix_rewrite"`
	Rewrites             []string `config:"rewrites"`
}

type Response struct {
	Headers map[string]string `config:"headers"`
}

type HealthCheck struct {
	Enable bool `config:"enable"`

	//
	Method string  `config:"method,default=GET"`
	Path   string  `config:"path,default=/health"`
	Status []int64 `config:"status,default=[200]"`

	// ok means health check is ok, ignore real check
	Ok bool `config:"ok"`
}

type Auth struct {
	Type string `config:"type"`

	// type: basic
	Username string `config:"username"`
	Password string `config:"password"`

	// type: bearer
	Token string `config:"token"`

	// type: jwt
	Secret string `config:"secret"`

	// type: oauth2
	Provider     string   `config:"provider"`
	ClientID     string   `config:"client_id"`
	ClientSecret string   `config:"client_secret"`
	RedirectURL  string   `config:"redirect_url"`
	Scopes       []string `config:"scopes"`

	// type: oidc

	// type: service
}
