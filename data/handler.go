package data

type BaseHandler struct {
	Profile string
	Name    string
}

func (h *BaseHandler) GetProfile() string {
	return h.Profile
}

func (h *BaseHandler) GetName() string {
	return h.Name
}
