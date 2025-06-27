package store

type Message struct {
	Channel string
	Data    []byte
}

type Store interface {
	Publish(msg Message) error
	Subscribe(channel string) (<-chan Message, error)
}
