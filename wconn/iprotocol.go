package wconn

type Handler interface {
	Response(client *Client, data []byte) error
}

type Protocol interface {
	OnClientRegister(client *Client) (closed bool)

	OnClientUnregister(client *Client)

	Handler
}
