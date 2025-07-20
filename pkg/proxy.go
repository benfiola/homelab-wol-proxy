package wolproxy

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

// Proxy contains the internal data and state for the wol proxy
type Proxy struct {
	Address      string
	Logger       *slog.Logger
	ReverseProxy *httputil.ReverseProxy
	Router       *chi.Mux
	WolHostname string
	WolMacAddress  net.HardwareAddr
}

// Opts are arguments used to construct the wol proxy
// See: [New]
type Opts struct {
	Address     string
	Backend     string
	Logger      *slog.Logger
	WolHostname string
	WolMacAddress string
}

// New constructs a wol-proxy.
// Returns an error on failure
func New(opts Opts) (*Proxy, error) {
	fail := func(err error) (*Proxy, error) {
		return nil, err
	}

	backend, err := url.Parse(fmt.Sprintf("http://%s", opts.Backend))
	if err != nil {
		return fail(err)
	}

	logger := opts.Logger
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}

	wolHostname := opts.WolHostname
	if wolHostname == "" {
		wolHostname = backend.Hostname()
	}

	wolMacAddress, err := net.ParseMAC(opts.WolMacAddress)
	if err != nil {
		return fail(err)
	}

	reverseProxy := httputil.NewSingleHostReverseProxy(backend)
	router := chi.NewRouter()

	proxy := Proxy{
		Address:      opts.Address,
		Logger:       logger,
		ReverseProxy: reverseProxy,
		Router:       router,
		WolHostname: wolHostname,
		WolMacAddress: wolMacAddress,
	}

	router.Get("/health", proxy.GetHealth)
	router.HandleFunc("/*", proxy.ProxyRequest)

	return &proxy, nil
}

// Runs the wol proxy in a blocking call.
// Returns an error on failure.
func (p *Proxy) Run() error {
	return http.ListenAndServe(p.Address, p.Router)
}

// Health check endpoint for the proxy
func (p *Proxy) GetHealth(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

// Gets an ip address for the given hostname.
// Returns an error on failure.
func (p *Proxy) GetIpAddress(hostname string) (net.IP, error) {
	fail := func(err error) (net.IP, error) {
		return nil, err
	}

	addrs, err := net.LookupHost(hostname)
	if err != nil {
		return fail(err)
	}

	if len(addrs) == 0 {
		return fail(fmt.Errorf("ip address for %s not found", hostname))
	}

	ip := net.ParseIP(addrs[0])
	if ip == nil {
		return fail(fmt.Errorf("invalid ip address %s", addrs[0]))
	}

	return ip, nil
}

// Builds a magic packet for the given mac address.
func (p *Proxy) CreateMagicPacket(mac net.HardwareAddr) []byte {
	packet := make([]byte, 102)
	for index := 0; index < len(packet); index += 1 {
		if index < 6 {
			packet[index] = 0xFF
			continue
		}
		macIndex := index % 6
		packet[index] = mac[macIndex]
	}
	return packet
}

// Sends a wake on lan packet containing the given mac address to the given ip address
// Returns an error on failure
func (p *Proxy) SendWakeOnLan(ip net.IP, mac net.HardwareAddr) error {
	packet := p.CreateMagicPacket(mac)

	udpaddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:9", ip.String()))
	if err != nil {
		return err
	}

	client, err := net.DialUDP("udp", nil, udpaddr)
	if err != nil {
		return err
	}
	defer client.Close()

	_, err = client.Write(packet)
	if err != nil {
		return err
	}

	return nil
}

// Endpoint that proxies requests to the backend.  Before doing so, will issue a wake-on-lan packet to the backend.
func (p *Proxy) ProxyRequest(w http.ResponseWriter, r *http.Request) {
	ip, err := p.GetIpAddress(p.WolHostname)
	if err != nil {
		render.Status(r, 500)
		render.PlainText(w, r, fmt.Sprintf("could not get ip address: %s", err.Error()))
		return
	}
	
	err = p.SendWakeOnLan(ip, p.WolMacAddress)
	if err != nil {
		render.Status(r, 500)
		render.PlainText(w, r, fmt.Sprintf("could not send wake-on-lan packet: %s", err.Error()))
	}

	p.ReverseProxy.ServeHTTP(w, r)
}
