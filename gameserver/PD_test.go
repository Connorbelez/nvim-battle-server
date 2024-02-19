package gameserver

import (
	"bytes"
	"fmt"
	"math/rand"
	"net"
	"testing"
	"time"
	// "vbattle-server/gameserver"
)

type MockConn struct {
	ReadBuffer    bytes.Buffer
	WriteBuffer   bytes.Buffer
	LocalAddress  net.Addr
	RemoteAddress net.Addr
	ReadDeadline  time.Time
	WriteDeadline time.Time
	Closed        bool
}

func NewMockConn(localAddr, remoteAddr net.Addr) net.Conn {
	return &MockConn{
		LocalAddress:  localAddr,
		RemoteAddress: remoteAddr,
	}
}

func NewRandMockConn() (*MockConn, net.Addr, net.Addr) {
	localAddr := &net.TCPAddr{
		IP:   net.IPv4(byte(rand.Intn(256)), byte(rand.Intn(256)), byte(rand.Intn(256)), byte(rand.Intn(256))),
		Port: rand.Intn(65535),
	}
	remoteAddr := &net.TCPAddr{
		IP:   net.IPv4(byte(rand.Intn(256)), byte(rand.Intn(256)), byte(rand.Intn(256)), byte(rand.Intn(256))),
		Port: rand.Intn(65535),
	}

	return &MockConn{
		LocalAddress:  localAddr,
		RemoteAddress: remoteAddr,
	}, localAddr, remoteAddr
}

func (mc *MockConn) Read(b []byte) (n int, err error) {
	if mc.Closed {
		return 0, net.ErrClosed
	}
	return mc.ReadBuffer.Read(b)
}

func (mc *MockConn) Write(b []byte) (n int, err error) {
	if mc.Closed {
		return 0, net.ErrClosed
	}
	return mc.WriteBuffer.Write(b)
}

func (mc *MockConn) Close() error {
	mc.Closed = true
	return nil
}

func (mc *MockConn) LocalAddr() net.Addr {
	return mc.LocalAddress
}

func (mc *MockConn) RemoteAddr() net.Addr {
	return mc.RemoteAddress
}

func (mc *MockConn) SetDeadline(t time.Time) error {
	mc.SetReadDeadline(t)
	mc.SetWriteDeadline(t)
	return nil
}

func (mc *MockConn) SetReadDeadline(t time.Time) error {
	mc.ReadDeadline = t
	return nil
}

func (mc *MockConn) SetWriteDeadline(t time.Time) error {
	mc.WriteDeadline = t
	return nil
}

// MockAddr implements net.Addr for testing purposes.
type MockAddr struct {
	NetType string
	Address string
}

// Network returns the network type (e.g., "tcp", "udp").
func (m *MockAddr) Network() string {
	return m.NetType
}

// String returns the address.
func (m *MockAddr) String() string {
	return m.Address
}

func TestResize_ElementsPreserved(t *testing.T) {
	// Setup initial queue with some elements, assuming a capacity of 4 to start
	initialCapacity := 4
	pDeque := PlayerDequeueFactory(initialCapacity)
	pDeque.SetE(3)
	pDeque.SetS(1)
	pDeque.SetL(2)

	mockAddr1 := MockAddr{
		NetType: "TCP",
		Address: "192.168.3.19",
	}
	// Mock clients (these could be more elaborate mocks or stubs as needed)
	mockAddr2 := MockAddr{
		NetType: "TCP",
		Address: "192.168.3.20",
	}

	mockConn1 := NewMockConn(&mockAddr1, &mockAddr1)
	mockConn2 := NewMockConn(&mockAddr2, &mockAddr2)

	mockClient1 := CreatePlayer(&mockConn1)
	mockClient2 := CreatePlayer(&mockConn2)

	pDeque.GetQ()[1] = mockClient1
	pDeque.GetQ()[2] = mockClient2

	// Act - Resize the queue
	pDeque.Resize()

	// Assert - Check if the capacity is doubled
	if gotCapacity := cap(pDeque.GetQ()); gotCapacity != 2*initialCapacity {
		t.Errorf("Resize() did not double the capacity, got = %v, want %v", gotCapacity, 2*initialCapacity)
	}

	t.Log(pDeque.GetS(), pDeque.GetE())
	t.Log(pDeque.GetQ()[0], pDeque.GetQ()[1], pDeque.GetQ()[2])
	// Assert - Check if the elements are preserved and in the correct order
	if pDeque.GetQ()[0] != mockClient1 || pDeque.GetQ()[1] != mockClient2 {
		t.Errorf("Resize() did not preserve elements in correct order")
	}
}
func TestPop_ModuloArithmatic(t *testing.T) {
	// Setup initial queue with some elements, assuming a capacity of 4 to start
	initialCapacity := 5
	pDeque := PlayerDequeueFactory(initialCapacity)
	conns := make([]*net.Conn, 10)
	for i := 0; i < 10; i++ {

		mockAddr1 := MockAddr{
			NetType: "TCP",
			Address: "192.168.3.1" + fmt.Sprint(i),
		}
		mockConn1 := NewMockConn(&mockAddr1, &mockAddr1)
		// mockPlayer := CreatePlayer(&mockConn1)
		// mockClient1 := CreatePlayer(&mockConn1)
		conns[i] = &mockConn1
	}

	pDeque.Accept(conns[0])
	pDeque.Accept(conns[1])
	pDeque.Accept(conns[2])

	if pDeque.GetQ()[0] == nil {
		t.Error("nil value")
	}
	if res, _ := pDeque.Get(0); res != pDeque.GetQ()[0] {
		t.Error("GET NOT WORKING", res, pDeque.GetQ()[0])
	}
	if res, _ := pDeque.Get(1); res != pDeque.GetQ()[1] {
		t.Error("GET NOT WORKING", res, pDeque.GetQ()[1])
	}
	if res, _ := pDeque.Get(2); res != pDeque.GetQ()[2] {
		t.Error("GET NOT WORKING", res, pDeque.GetQ()[2])
	}

	//UNDER: 0, 1, 2, 3 Logic: 0, 1, 2, 3
	t.Log(pDeque.GetQ())

	res1, _ := pDeque.PopLeft()
	t.Log(pDeque.GetQ())
	res2, _ := pDeque.PopLeft()
	t.Log(pDeque.GetQ())

	res3, _ := pDeque.PopLeft()
	t.Log(pDeque.GetQ())

	pDeque.Accept(conns[3])
	t.Log(pDeque.GetQ())
	if res, _ := pDeque.Get(0); res != pDeque.GetQ()[3] {
		t.Error("GET NOT WORKING", res, pDeque.GetQ()[3])
	}
	t.Log(pDeque.GetQ())
	//UNDER nil, nil, nil, 3, nil | Logic: 3
	if res, _ := pDeque.Get(0); res == nil || res != pDeque.GetQ()[3] {
		t.Log(pDeque.GetQ())
		t.Error("GET NOT WORKING", res, pDeque.GetQ()[3])
	}

	temp := pDeque.GetQ()[3]

	t.Log(res1, res2, res3)

	//Pushed 4, popped 3,
	pDeque.Accept(conns[4])
	t.Log(pDeque.GetQ())
	//UNDER nil, nil, nil, 3, 4 | Logic: 3, 4
	pDeque.Accept(conns[5]) //get(2), index[5]
	t.Log(pDeque.GetQ())
	//UNDER 5, nil, nil, 3, 4 | Logic: 3, 4, 5

	arrLastRes := pDeque.GetQ()[0]
	getLastRes, _ := pDeque.Get(2)

	if arrLastRes == nil || getLastRes == nil || arrLastRes != getLastRes {
		t.Error("issue with last element arith", getLastRes, arrLastRes)
	}

	for i, v := range pDeque.GetQ() {
		t.Log(i, v)
	}

	if res, _ := pDeque.Get(0); res != temp {
		t.Error("Get(0): ", res, " temp: ", temp)
	}

	if getLastRes != arrLastRes && arrLastRes != nil {
		t.Error("issue with last element arith", getLastRes, arrLastRes)
	}

	pDeque.Accept(conns[5])
	pDeque.Accept(conns[6])
	if res, _ := pDeque.Get(0); res != temp && temp != nil && res != nil {
		t.Error("Last val not expected", res, temp)
	}

}

func TestMatchMake_WithRealQueue(t *testing.T) {
	// Setup mock addresses for local and remote connections
	localAddr := &MockAddr{NetType: "tcp", Address: "127.0.0.1:8080"}
	remoteAddr1 := &MockAddr{NetType: "tcp", Address: "127.0.0.1:8081"}
	remoteAddr2 := &MockAddr{NetType: "tcp", Address: "127.0.0.1:8082"}

	// Create mock connections for two players
	conn1 := NewMockConn(localAddr, remoteAddr1)
	conn2 := NewMockConn(localAddr, remoteAddr2)

	// Initialize the PlayerDequeue (PDeque)
	pDeque := PlayerDequeueFactory(4)

	// Add mock clients to the queue
	_, err := pDeque.Accept(&conn1)
	if err != nil {
		t.Fatalf("Failed to accept connection 1: %v", err)
	}
	_, err = pDeque.Accept(&conn2)
	if err != nil {
		t.Fatalf("Failed to accept connection 2: %v", err)
	}

	// Ensure the queue has the expected number of clients
	if gotSize := pDeque.GetSize(); gotSize != 2 {
		t.Fatalf("Expected queue size of 2, got %d", gotSize)
	}

	// Initialize GameMaster with the real queue
	gameMaster := GameMaster{
		PlayerQ:     pDeque,
		GameLobbies: make(map[string]*GameInstance),
	}

	// Perform matchmaking
	gameMaster.MatchMake()

	// Assert that a game lobby is created
	if len(gameMaster.GameLobbies) != 1 {
		t.Errorf("Expected 1 game lobby to be created, got %d", len(gameMaster.GameLobbies))
	}

	for k, v := range gameMaster.GameLobbies {
		t.Log(k, v)
		t.Log(v.playersReady, v.running, v.Players)
		t.Log(v.PlayerState)
	}
	// Assert that the game lobby has the expected number of players

}
