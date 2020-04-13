package encoder

type MsgProto interface {
	Marshal(interface{}) ([]byte, error)

	Unmarshal([]byte, interface{}) error
}
