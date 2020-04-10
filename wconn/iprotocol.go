package wconn

type Protocol interface {
	OnClientRegister(client *Client) (closed bool)

	OnClientUnregister(client *Client)

	Response(client *Client, data []byte) error

	SendAckMsg(from, target int64, msg string)
}
