package utils


import (
	"github.com/google/uuid"
	"net"
	"fmt"
)

type Client interface {
	Send(message string) error
	Receive() (string, error)
	Close() error
}
func Test(){
	fmt.Println("SUC")
}
type SocketClient struct {
	ID         string
	Socket     *net.Conn
	ClientAddr string
}

func CreatePlayer(c net.Conn) *SocketClient {
	t := SocketClient{ID: uuid.NewString(), Socket: &c, ClientAddr: c.LocalAddr().String()}
	return &t
}

func (sc *SocketClient) Send(message string) (int, error) {
	// Implementation to send a message through the socket
	n, err := (*sc.Socket).Write([]byte(message))
	return n, err
}

func (sc *SocketClient) Receive() (string, int, error) {
	// Implementation to receive a message from the Socket
	outBuf := make([]byte, 0, 1024)
	n, err := (*sc.Socket).Read(outBuf)

	return string(outBuf), n, err
}

func (sc *SocketClient) Close() error {
	// Implementation to close the socket connection
	(*sc).Close()
	return nil
}

