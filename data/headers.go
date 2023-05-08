package data

const (
	HEADER_X_FAULTMONKEY_API_CUSTOMER = "X-Faultmonkey-Api-Customer"
	HEADER_X_FAULTMONKEY_API_KEY      = "X-Faultmonkey-Api-Key"
	HEADER_X_FAULTMONKEY_ERROR        = "X-Faultmonkey-Error"
	HEADER_X_FAULTMONKEY_TAG          = "X-Faultmonkey-Tag"

	// This header is set by the broker to indicate the requested
	// upstream service
	HEADER_X_FAULTMONKEY_UPSTREAM = "X-Faultmonkey-Upstream"

	// This is a response header that indicates the backend
	// which handled the request
	HEADER_X_FAULTMONKEY_BACKEND = "X-Faultmonkey-Backend"

	// This is a response header that indicates the round-trip
	// latency of the backend that handled the request
	// (in millis)
	HEADER_X_FAULTMONKEY_BACKEND_LATENCY = "X-Faultmonkey-Backend-Latency"

	// If clients want to send particular request IDs (e.g.
	// to help with tracing) then this is the request header
	// to use.
	//
	// If it is provided, then it is also returned in the response
	HEADER_X_REQUEST_ID = "X-Request-Id"
)
