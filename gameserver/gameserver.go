package gameserver

import (
	// "crypto"
	// "vbattle-server/utils"
	// "encoding"
	// "container/heap"
	// "container/dequeu"
	"context"
	"errors"
	"fmt"
	"net"
	"sync"

	// "strconv"
	"github.com/google/uuid"
	// "strconv"
	// "time"
)

func Start() {
	// Listen for incoming connections on port 8080
	ln, err := net.Listen("tcp", ":80")
	// fmt.Println(ln.Addr().String())
	if err != nil {
		fmt.Println(err)
		return
	}

	// Accept incoming connections and handle them
	for {
		print("waiting...")
		conn, err := ln.Accept()
		print("ACCEPTED!")
		// c := utils.CreatePlayer(conn)
		// print(conn)
		if err != nil {
			fmt.Println(err)
			continue
		}

		// Handle the connection in a new goroutine
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	// Close the connection when we're done
	defer conn.Close()

	keepOpen := true
	fmt.Println(conn)
	for keepOpen {

		// Read incoming data
		buf := make([]byte, 1024)
		_, err := conn.Read(buf)

		if err != nil {
			fmt.Println(err)
			return
		}
		// clientTimeStamp, err := strconv.ParseInt(buf[:len(buf)-1])

		// serverTime := time.Now()
		// fmt.Printf("TIME: %v", serverTime)
		// Print the incoming data
		fmt.Printf("Received: %s", buf)
		// res := "from server" + (serverTime.Format("15:04:05.00"))
		conn.Write([]byte("SERVER MESSAGE\n"))
		print("DONE WRITING")
	}
}

// type PlayerQ interface {
// 	Accept(conn *net.Conn)
//
// 	PopLeft() Client
//
// 	// PopIndex(i int) Client
//
// 	PopClient(id string) *Client
//
// 	GetClient(id string) *Client
//
// 	MatchMake()
// }

// FUCK LRU ALL MY HOMIES LOVE DEQUES
// Consider re-using these
// type PlayerNode struct {
// 	head   *PlayerNode
// 	tail   *PlayerNode
// 	next   *PlayerNode
// 	prev   *PlayerNode
// 	client *Client
// }

//	type GameState struct {
//		active    bool
//		connected bool
//	}
//
//	type GameLobby struct {
//		p1 Client
//		p2 Client
//	  state GameState
//	}
type Player struct {
	ID          string
	client      Client
	ctx         context.Context
	cancelFunc  context.CancelFunc // Used to cancel the context
	eventBuffer []string
}

func NewPlayer(c Client) *Player {
	ctx, cancelFunc := context.WithCancel(context.Background())
	return &Player{
		ID:          uuid.NewString(), // Implement this function to generate unique IDs
		client:      c,
		ctx:         ctx,
		cancelFunc:  cancelFunc,
		eventBuffer: make([]string, 100),
	}
}

func (p *Player) ReadActions(actionCh chan<- string, errCh chan<- error) {
	go func() {
		buffer := make([]byte, 1024)
		for {
			select {
			case <-p.ctx.Done():
				// Context was cancelled, stop the goroutine
				return
			default:
				// Attempt to read from the connection
				n, err := p.client.Receive(buffer)
				if err != nil {
					errCh <- err
					return
				}
				p.eventBuffer = append(p.eventBuffer, string(buffer[:n]))
				actionCh <- string(buffer[:n])
			}
		}
	}()
}

func (gm *GameMaster) StartGame(game *GameInstance) {
	actionCh := make(chan string)
	errCh := make(chan error)

	// Start reading actions for each player
	for _, player := range game.Players {
		player.ReadActions(actionCh, errCh)
	}

	go func() {
		for {
			select {
			case action := <-actionCh:
				print("ACTION", action)
				// Process action
			case <-errCh:
				// One of the players disconnected, handle cleanup
				for _, player := range game.Players {
					player.cancelFunc() // This cancels the context, stopping ReadActions goroutine
				}
				return // Exit the game loop
			}
		}
	}()
}

func (gm *GameMaster) MatchMake() {
	for gm.PlayerQ.GetSize() > 1 {
		p1, err := gm.PlayerQ.PopLeft()
		p2, err := gm.PlayerQ.PopLeft()

		p1State := make(map[string]string)
		p2State := make(map[string]string)

		// players :=

	}
}

type GameInstance struct {
	Players      []Player
	running      bool
	playersReady int
	PlayerState  map[string]map[string]string
}

type GameMaster struct {
	PlayerQ     PlayerDequeue
	GameLobbies map[string]*GameInstance
	lock        sync.RWMutex
}

type PlayerDequeue interface {
	Get(i int) (Client, error)
	Set(i int, c Client) error
	// PushLeft(*Client)
	// PushRight(*Client)
	Accept(c net.Conn) error
	PopLeft() (Client, error)
	GetSize() int
	Resize()
}

type PDeque struct {
	Q []Client
	s int
	e int
}

func (q PDeque) GetSize() int {
	return q.e - q.s
}

func (PQ PDeque) Get(i int) (Client, error) {
	if i < 0 || i > PQ.e-1 {
		return nil, errors.New("skill issue")
	}
	return PQ.Q[(i+PQ.e)%len(PQ.Q)], errors.New("skill issue")
}

func (PQ PDeque) Set(i int, c Client) error {
	if i < 0 || i > PQ.e-1 {
		return errors.New("skill issue")
	}
	return nil
}

func (PQ PDeque) PopLeft() (Client, error) {
	if len(PQ.Q) < 1 {
		return nil, errors.New("skill issue")
	}
	t := PQ.Q[PQ.s]
	// PQ.Q[PQ.s] = nil
	PQ.s += 1
	return t, nil
}

func (PQ PDeque) MatchMake()

func (PQ PDeque) Accept(c *net.Conn) (int, error) {
	if !(PQ.e < len(PQ.Q)-1) {
		PQ.Resize()
	}
	PQ.Q[PQ.e] = CreatePlayer(c)
	return PQ.e - PQ.s, nil

}

func (PQ PDeque) Resize() {
	t := make([]Client, 2*len(PQ.Q), 2*len(PQ.Q))
	copy(t, PQ.Q)
	PQ.Q = t
}

type Client interface {
	Send(message string) error
	Receive([]byte) (int, error)
	Close() error
}

// Assuming the rest of your implementation is correct
type SocketClient struct {
	ID         string
	Socket     *net.Conn
	ClientAddr string
}

func CreatePlayer(c *net.Conn) Client { // Return Client interface directly
	t := SocketClient{ID: uuid.NewString(), Socket: c, ClientAddr: (*c).LocalAddr().String()}
	return &t // This now correctly returns something that implements Client
}

func (sc *SocketClient) Send(message string) error { // Only return error
	n, err := (*sc.Socket).Write([]byte(message))
	fmt.Println(n) // Assuming you want to do something with n, or else remove this
	return err
}

func (sc *SocketClient) Receive(outBuf []byte) (int, error) {
	// outBuf := make([]byte, 1024) // Allocate buffer with initial size
	n, err := (*sc.Socket).Read(outBuf)
	return n, err
	// return string(outBuf[:n]), n, err // Ensure to slice outBuf up to n
}

func (sc *SocketClient) Close() error {
	return (*sc.Socket).Close() // Correctly close the net.Conn
}
