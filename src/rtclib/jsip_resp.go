// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//

// JSIP Response

package rtclib

// JSIPRespType
type JSIPResponseType int

const (
	JSIPReq JSIPResponseType = iota + 1

	// 1xx: Provisional -- request received, continuing to process the request;
	JSIPProvisionalResp

	// 2xx: Success -- the action was successfully received, understood, and accepted;
	JSIPSuccessResp

	// 3xx: Redirection -- further action needs to be taken in order to complete the request;
	JSIPRedirectionResp

	// 4xx: Client Error -- the request contains bad syntax or cannot be fulfilled at this server;
	JSIPClientErrResp

	// 5xx: Server Error -- the server failed to fulfill an apparently valid request;
	JSIPServerErrResp

	// 6xx: Global Failure -- the request cannot be fulfilled at any server.
	JSIPGlobalFailureResp
)

var jsipRespDesc = map[int]string{
	int(JSIPProvisionalResp):   "User Defined Provisional Response",
	int(JSIPSuccessResp):       "User Defined Success Response",
	int(JSIPRedirectionResp):   "User Defined Redirection Response",
	int(JSIPClientErrResp):     "User Defined Client Error Response",
	int(JSIPServerErrResp):     "User Defined Server Error Response",
	int(JSIPGlobalFailureResp): "User Defined Global Failure Response",
	100: "Trying",
	180: "Ringing",
	181: "Call Is Being Forwarded",
	182: "Queued",
	183: "Session Progress",
	200: "OK",
	202: "Accepted",
	300: "Multiple Choices",
	301: "Moved Permanently",
	302: "Moved Temporarily",
	305: "Use Proxy",
	380: "Alternative Service",
	400: "Bad Request",
	401: "Unauthorized",
	402: "Payment Required",
	403: "Forbidden",
	404: "Not Found",
	405: "Method Not Allowed",
	406: "Not Acceptable",
	407: "Proxy Authentication Required",
	408: "Request Timeout",
	410: "Gone",
	413: "Request Entity Too Large",
	414: "Request-URI Too Large",
	415: "Unsupported Media Type",
	416: "Unsupported URI Scheme",
	420: "Bad Extension",
	421: "Extension Required",
	423: "Interval Too Brief",
	480: "Temporarily not available",
	481: "Call Leg/Transaction Does Not Exist",
	482: "Loop Detected",
	483: "Too Many Hops",
	484: "Address Incomplete",
	485: "Ambiguous",
	486: "Busy Here",
	487: "Request Terminated",
	488: "Not Acceptable Here",
	491: "Request Pending",
	493: "Undecipherable",
	500: "Internal Server Error",
	501: "Not Implemented",
	502: "Bad Gateway",
	503: "Service Unavailable",
	504: "Server Time-out",
	505: "SIP Version not supported",
	513: "Message Too Large",
	600: "Busy Everywhere",
	603: "Decline",
	604: "Does not exist anywhere",
	606: "Not Acceptable",
}

// Get JSIP Response type by response code
func JSIPRespType(code int) JSIPResponseType {
	if code == 0 {
		return JSIPReq
	}

	if code/100 == 1 {
		return JSIPProvisionalResp
	}

	if code/100 == 2 {
		return JSIPSuccessResp
	}

	if code/100 == 3 {
		return JSIPRedirectionResp
	}

	if code/100 == 4 {
		return JSIPClientErrResp
	}

	if code/100 == 5 {
		return JSIPServerErrResp
	}

	if code/100 == 6 {
		return JSIPGlobalFailureResp
	}

	return JSIPResponseType(Unknown)
}

// Get JSIP Response description by response code
func JSIPRespDesc(code int) string {
	output := jsipRespDesc[code]

	if output == "" {
		return jsipRespDesc[int(JSIPRespType(code))]
	}

	return output
}
