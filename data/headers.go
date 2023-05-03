package data

const (
	HEADER_X_FAULTMONKEY_API_CUSTOMER = "X-Faultmonkey-Api-Customer"
	HEADER_X_FAULTMONKEY_API_KEY      = "X-Faultmonkey-Api-Key"
	HEADER_X_ATTENUATOR_ERROR     = "X-Attenuator-Error"
	HEADER_X_FAULTMONKEY_TAG          = "X-faultmonkey-tag"

	// This is a response header that indicates the backend
	// which handled the request
	HEADER_X_FAULTMONKEY_BACKEND = "X-faultmonkey-backend"

	// If clients want to send particular request IDs (e.g.
	// to help with tracing) then this is the request header
	// to use.
	//
	// If it is provided, then it is also returned in the response
	HEADER_X_REQUEST_ID = "X-request-id"
)
