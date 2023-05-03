package data

type Server struct {
	Name   string `yaml:"name" json:"name"`
	Listen string `yaml:"listen" json:"listen"`
	Enable bool   `yaml:"enable" json:"enable"`

	// Mapping of host header value -> implementation
	Hosts map[string]*ServerHost
}

type ServerHost struct {
	PathologyProfileName string `yaml:"pathology" json:"pathology"`

	// The actual pathology profile instance gets backpatched
	pathologyProfile PathologyProfile
}

func (s *ServerHost) GetPathologyProfile() PathologyProfile {
	return s.pathologyProfile
}
