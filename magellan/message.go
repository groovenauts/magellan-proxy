package magellan

type RequestMessage struct {
	V   int `codec:"v" msgpack:"v"`
	Env struct {
		ReferenceIp      string `codec:"REFERENCE_IP", msgpack:"REFERENCE_IP"`
		Method           string `codec:"METHOD", msgpack:"METHOD"`
		Url              string `codec:"URL", msgpack:"URL"`
		OauthRequesterId string `codec:"OAUTH_REQUESTER_ID", msgpack:"OAUTH_REQUESTER_ID"`
		ContentType      string `codec:"CONTENT_TYPE", msgpack:"CONTENT_TYPE"`
		ServerName       string `codec:"SERVER_NAME", msgpack:"SERVER_NAME"`
		ServerPort       int    `codec:"SERVER_PORT", msgpack:"SERVER_PORT"`
		PathInfo         string `codec:"PATH_INFO", msgpack:"PATH_INFO"`
		QUERY_STRING     string `codec:"QUERY_STRING", msgpack:"QUERY_STRING"`
	} `codec:"env" msgpack:"env"`
	Headers map[string]string      `codec:"headers", msgpack:"headers"`
	Options map[string]interface{} `codec:"options", msgpack:"options"`
}

type ResponseMessage struct {
	Headers      map[string]string `codec:"headers", msgpack:"headers"`
	Status       string            `codec:"status", msgpack:"status"`
	Body         string            `codec:"body", msgpack:"body"`
	BodyEncoding string            `codec:"body_encoding", msgpack:"body_encoding"`
}

// vim:set noexpandtab ts=2:
