package broker

type Upstream struct {
	Name     string    `json:"name"`
	Rule     string    `json:"rule"`
	Backends []Backend `json:"backends"`
}
