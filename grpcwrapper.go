package grpcwrapper

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"time"
)

type Conn struct {
	conn *grpc.ClientConn

	Context context.Context
	cancel  context.CancelFunc

	lastActive time.Time
	factory    func() (*grpc.ClientConn, error)
	client     func(*grpc.ClientConn) (interface{})

	client_    interface{}
}

func (g *Conn) Client() interface{} {
	if g.client_ == nil {
		g.client_ = g.client(g.conn)
	}

	return g.client_
}

func CreateConn(factory func()(*grpc.ClientConn,error),
	client func(*grpc.ClientConn)(interface{})) *Conn {
	return &Conn{
		lastActive: time.Now(),
		factory: factory,
		client: client,
	}
}

func (g *Conn) Reconnect() error {
	if g.conn != nil && g.lastActive.Add(5*time.Second).Before(time.Now()) {
		_ = g.conn.Close()
		g.conn = nil
	}

	if g.conn == nil {
		var err error
		g.conn, err = g.factory()
		if err != nil {
			g.conn = nil
			return err
		}
	}

	g.lastActive = time.Now()
	return nil
}

func (w *Conn) Cancel() {
	if w == nil {
		return
	}

	if w.cancel != nil {
		w.cancel()
	}

	// if w.conn != nil {
	// 	w.conn.Close()
	// }
}

func (g *Conn) Clone() (*Conn, error) {
	err := g.Reconnect()
	if err != nil {
		return nil, err
	}

	gctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	return &Conn{
		conn:       g.conn,
		Context:    gctx,
		cancel:     cancel,
		lastActive: g.lastActive,
		factory:    g.factory,
		client:     g.client,
		client_:    g.client(g.conn),
	}, nil
}
