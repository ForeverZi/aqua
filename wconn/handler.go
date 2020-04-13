package wconn

type EchoHanlder struct{}

func (echo *EchoHanlder) Response(client *Client, data []byte) error {
	client.Send(data)
	return nil
}
