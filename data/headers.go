package data

const (
	HEADER_X_MIGALOO_API_CUSTOMER = "X-Migaloo-Api-Customer"
	HEADER_X_MIGALOO_API_KEY      = "X-Migaloo-Api-Key"
	HEADER_X_ATTENUATOR_ERROR     = "X-Attenuator-Error"
	HEADER_X_MIGALOO_TAG          = "X-migaloo-tag"

	// This is a response header that indicates the backend
	// which handled the request
	HEADER_X_MIGALOO_BACKEND = "X-migaloo-backend"

	// If clients want to send particular request IDs (e.g.
	// to help with tracing) then this is the request header
	// to use.
	//
	// If it is provided, then it is also returned in the response
	HEADER_X_REQUEST_ID = "X-request-id"
)
