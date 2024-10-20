package peer

// Pool represents a peer client pool over groups of peers.
// It handles operations over groups of clients.
type Pool struct {
	clients []*Client
}

func NewPool(clients []*Client) *Pool {
	return &Pool{clients: clients}
}

// Add adds a client to the pool, if it does not already exist.
func (p *Pool) Add(client *Client) {
	for _, c := range p.clients {
		if c.String() == client.String() {
			return
		}
	}
	p.clients = append(p.clients, client)
}

// For temporary backwards compatibility.
func (p *Pool) GetClients() []*Client {
	return p.clients
}
