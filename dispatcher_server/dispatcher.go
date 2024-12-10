package dispatcher

type Dispatcher struct {
	baseHost string
	tunnels  map[string]*Tunnel
}

func New(baseHost string) Dispatcher {
	return Dispatcher{
		baseHost: baseHost,
		tunnels:  make(map[string]*Tunnel),
	}
}
