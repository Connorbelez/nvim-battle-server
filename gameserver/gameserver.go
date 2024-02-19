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

type TwoWayConn struct {
	inbound  *net.Conn
	outbound *net.Conn
}

func Start() {
	GM := GameMasterFactory(100)
	// Listen for incoming connections on port 8080
	ln, err := net.Listen("tcp", ":80")
	// fmt.Println(ln.Addr().String())
	if err != nil {
		fmt.Println(err)
		return
	}
	conMap := make(map[string]TwoWayConn)

	// Accept incoming connections and handle them
	for {
		print("waiting...")
		conn, err := ln.Accept()
		print(conn.RemoteAddr().String())
		// GM.PlayerQ.Accept(&conn)
		// if GM.PlayerQ.GetSize() > 1 {
		// 	GM.MatchMake()
		// }

		print("ACCEPTED!")
		// c := utils.CreatePlayer(conn)
		// print(conn)
		if err != nil {
			fmt.Println(err)
			continue
		}
		// recieve message from client

		buf := make([]byte, 1024)
		n, err := conn.Read(buf)

		if err != nil {

			fmt.Println(err)
		}
		// parse id and role in form of "R:ID"
		splitStr := string(buf[:n])
		//Parse string split at ":"
		role := splitStr[:2]
		id := splitStr[2:]
		print(role, id)
		// if id is not in map, add it
		// rec := false
		if _, ok := conMap[id]; !ok {
			if role == "R:" {
				print("got first R")
				// rec = true
				conMap[id] = TwoWayConn{inbound: &conn, outbound: nil}
			} else {
				print("got first S")

				conMap[id] = TwoWayConn{inbound: nil, outbound: &conn}
			}
		} else {
			temp := conMap[id]
			if role == "R:" {
				// rec = true
				print("got second R")
				temp.inbound = &conn
			} else {
				print("got second S")
				temp.outbound = &conn
			}
			conMap[id] = temp
		}
		// Send ok message to client

		//check if the twowaycon is complete
		if conMap[id].inbound != nil && conMap[id].outbound != nil {
			fmt.Println("Two way connection complete")
			conn.Write([]byte("READY\n"))
			//push to queue
			GM.PlayerQ.Accept(conMap[id].inbound, conMap[id].outbound)
			if GM.PlayerQ.GetSize() > 1 {
				GM.MatchMake()
			}

		} else {
			fmt.Println("Two way connection incomplete")
			conn.Write([]byte("WAIT\n"))
		}

		// Handle the connection in a new goroutine
		// if !rec {
		// 	fmt.Println("Not a reciver, handling")
		// 	go handleConnection(conn, conMap, id, role)
		// }

	}
}

func handleConnection(conn net.Conn, conMap map[string]TwoWayConn, id string, role string) {
	// Close the connection when we're done
	defer conn.Close()
	fmt.Println("HANDLING :", role)

	keepOpen := true
	fmt.Println(conn)

	for keepOpen {
		fmt.Println("HANDLING :", role)
		print(conMap)

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
		fmt.Printf("Received: %s\n", buf)

		// res := "from server" + (serverTime.Format("15:04:05.00"))
		serverRes := "SERVER MESSAGE:" + string(buf) + "\n"
		// if the twoWayCon struct has an outbound connection, write to it
		fmt.Println("ID: ", id)
		fmt.Println("CONMAP: ", conMap[id])
		if _, ok := conMap[id]; ok {
			fmt.Println("OUT: ", conMap[id].outbound)

			fmt.Println("IN: ", conMap[id].inbound)
			if conMap[id].outbound != nil {
				fmt.Println("Writing to outbound")
				// (*conMap[id].outbound).Write([]byte(serverRes))
				out := *conMap[id].outbound
				in := *conMap[id].inbound
				fmt.Println("OUT ADDR: ", out.RemoteAddr().String())
				fmt.Println("IN ADDR: ", in.RemoteAddr().String())
				// out.Write([]byte(serverRes))
				in.Write([]byte(serverRes))
				print("DONE WRITING: ", serverRes)
			}
		}

		// conn.Write([]byte(serverRes))
		// print("DONE WRITING: ", serverRes)
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

func (p *Player) RecieveAction(actionCh chan<- string, errCh chan<- error) {
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
	// Create channels for each player
	fmt.Println("STARTING GAME")
	p1ActionCh := make(chan string)
	p1ErrCh := make(chan error)
	p2ActionCh := make(chan string)
	p2ErrCh := make(chan error)

	// Send "READY UP!" message to both players
	// game.Players[0].client.Send("READY UP!")
	// game.Players[1].client.Send("READY UP!")

	// Start reading actions for each player
	game.Players[0].RecieveAction(p1ActionCh, p1ErrCh)
	game.Players[1].RecieveAction(p2ActionCh, p2ErrCh)

	// Wait for confirmation from both players
	// for i := 0; i < 2; i++ {
	// 	select {
	// 	case <-p1ActionCh:
	// 		// Confirmation from player 1
	// 	case <-p2ActionCh:
	// 		// Confirmation from player 2
	// 	case <-p1ErrCh:
	// 		// Error from player 1, handle cleanup
	// 		game.Players[0].cancelFunc()
	// 		return
	// 	case <-p2ErrCh:
	// 		// Error from player 2, handle cleanup
	// 		game.Players[1].cancelFunc()
	// 		return
	// 	}
	// }

	// Relay events from one player to the other
	go func() {
		for action := range p1ActionCh {
			// Relay action from player 1 to player 2
			fmt.Println("RELAYING ACTION", action)
			game.Players[1].client.Send(action)
		}
	}()

	go func() {
		for action := range p2ActionCh {
			// Relay action from player 2 to player 1
			fmt.Println("RELAYING ACTION", action)
			game.Players[0].client.Send(action)
		}
	}()

	// Handle errors in a separate goroutine
	go func() {
		for {
			select {
			case <-p1ErrCh:
				// Error from player 1, handle cleanup
				game.Players[0].cancelFunc()
				return
			case <-p2ErrCh:
				// Error from player 2, handle cleanup
				game.Players[1].cancelFunc()
				return
			}
		}
	}()
}

type GameInstance struct {
	Players      []*Player
	running      bool
	playersReady int
	PlayerState  []map[string][]string
}

func (gm *GameMaster) MatchMake() {
	fmt.Println("MATCHMAKING")
	for gm.PlayerQ.GetSize() > 1 {
		p1, _ := gm.PlayerQ.PopLeft()
		p2, _ := gm.PlayerQ.PopLeft()

		p1State := make(map[string][]string)
		p2State := make(map[string][]string)

		players := []*Player{NewPlayer(p1), NewPlayer(p2)}

		gi := GameInstance{
			Players:      players,
			running:      false,
			playersReady: 0,
			PlayerState: []map[string][]string{
				p1State,
				p2State,
			},
		}

		gm.GameLobbies[uuid.NewString()] = &gi
		go gm.StartGame(&gi)

	}
}

func GameMasterFactory(n int) *GameMaster {
	return &GameMaster{
		PlayerQ:     PlayerDequeueFactory(n),
		GameLobbies: make(map[string]*GameInstance),
	}
}

type GameMaster struct {
	PlayerQ     PlayerDequeue
	GameLobbies map[string]*GameInstance
	lock        sync.RWMutex
}

type PlayerDequeue interface {
	Get(i int) (Client, error)
	Set(i int, c Client) error
	GetQ() []Client
	Accept(cin *net.Conn, cout *net.Conn) (int, error)
	PopLeft() (Client, error)
	GetSize() int
	Resize()
	GetS() int
	GetE() int
	SetS(int)
	SetE(int)
	SetL(int)
}

func PlayerDequeueFactory(n int) PlayerDequeue {
	return &PDeque{
		Q: make([]Client, n),
		S: 0,
		E: 0,
		L: 0,
	}
}

type PDeque struct {
	Q []Client
	S int
	E int
	L int
}

func (p *PDeque) SetS(s int) {
	p.S = s
}

func (p *PDeque) SetE(e int) {
	p.E = e
}

func (p *PDeque) SetL(l int) {
	p.L = l
}

func (p *PDeque) GetS() int {
	return p.S
}

func (p *PDeque) GetE() int {
	return p.E
}

func (q *PDeque) GetQ() []Client {
	return q.Q
}

func (q *PDeque) GetSize() int {
	return q.L
}

func (PQ *PDeque) Get(i int) (Client, error) {
	if i < 0 || i > PQ.L {
		return nil, errors.New("skill issue")
	}
	return PQ.Q[(i+PQ.S)%cap(PQ.Q)], nil
}

func (PQ *PDeque) Set(i int, c Client) error {
	if i < 0 || i > PQ.L {
		return errors.New("skill issue")
	}
	PQ.Q[(i+PQ.S)%PQ.L] = c

	return nil
}

func (PQ *PDeque) PopLeft() (Client, error) {

	if len(PQ.Q) < 1 {
		return nil, errors.New("skill issue")
	}
	t := PQ.Q[PQ.S]
	PQ.Q[PQ.S] = nil
	PQ.S += 1
	PQ.L -= 1

	return t, nil
}

func (PQ *PDeque) Accept(cin *net.Conn, cout *net.Conn) (int, error) {
	if !(PQ.L < cap(PQ.Q)) {
		PQ.Resize()
	}

	PQ.Q[PQ.E] = CreatePlayer(cin, cout)
	PQ.L += 1
	PQ.E = (PQ.E + 1) % cap(PQ.Q)
	return PQ.L, nil

}

func (PQ *PDeque) Resize() {
	t := make([]Client, 2*cap(PQ.Q), 2*cap(PQ.Q))
	// r := PQ.Q[PQ.s:]
	print(cap(t))
	print(cap(PQ.Q))
	if PQ.E < PQ.S {
		copy(t, PQ.Q[PQ.S:])
		copy(t, PQ.Q[:PQ.E])
	} else {
		copy(t, PQ.Q[PQ.S:PQ.E])
	}
	PQ.S = 0
	PQ.E = PQ.L
	PQ.Q = t

}

type Client interface {
	Send(message string) error
	Receive([]byte) (int, error)
	Close() (error, error)
}

// Assuming the rest of your implementation is correct
type SocketClient struct {
	ID            string
	InSocket      *net.Conn
	InClientAddr  string
	OutSocket     *net.Conn
	OutClientAddr string
}

func CreatePlayer(cin *net.Conn, cout *net.Conn) Client { // Return Client interface directly
	t := SocketClient{
		ID:            uuid.NewString(),
		InSocket:      cout,
		InClientAddr:  (*cin).LocalAddr().String(),
		OutSocket:     cin,
		OutClientAddr: (*cout).LocalAddr().String(),
	}
	return &t // This now correctly returns something that implements Client
}

func (sc *SocketClient) Send(message string) error { // Only return error
	n, err := (*sc.OutSocket).Write([]byte(message))
	fmt.Println(n) // Assuming you want to do something with n, or else remove this
	return err
}

func (sc *SocketClient) Receive(outBuf []byte) (int, error) {
	// outBuf := make([]byte, 1024) // Allocate buffer with initial size
	n, err := (*sc.InSocket).Read(outBuf)
	return n, err
	// return string(outBuf[:n]), n, err // Ensure to slice outBuf up to n
}

func (sc *SocketClient) Close() (error, error) {
	e1 := (*sc.InSocket).Close() // Correctly close the net.Conn
	e2 := (*sc.OutSocket).Close()
	return e1, e2 // Correctly close the net.Conn
}
