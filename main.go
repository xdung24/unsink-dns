package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/kardianos/service"
	"github.com/miekg/dns"
)

var (
	cache      = make(map[string]CacheEntry)
	cacheMutex sync.RWMutex
	// Using Cloudflare as the Upstream DoT server
	upstreamServer = "1.1.1.1:853"
	serverName     = "one.one.one.one"
	server         *dns.Server
	hostsMap       map[string][]net.IP
)

// CacheEntry stores the DNS message and its expiration time
type CacheEntry struct {
	Msg    *dns.Msg
	Expiry time.Time
}

type program struct{}

func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
}

func (p *program) run() {
	loadHosts()
	dns.HandleFunc(".", handleDNSRequest)
	server = &dns.Server{Addr: ":53", Net: "udp"}
	fmt.Printf("Starting DNS-over-TLS Proxy on %s...\n", server.Addr)

	err := server.ListenAndServe()
	if err != nil {
		log.Printf("Failed to start server: %s\n", err.Error())
	}
}

func (p *program) Stop(s service.Service) error {
	if server != nil {
		return server.Shutdown()
	}
	return nil
}

func loadHosts() {
	hostsMap = make(map[string][]net.IP)
	var path string
	if runtime.GOOS == "windows" {
		path = `C:\Windows\System32\drivers\etc\hosts`
	} else {
		path = "/etc/hosts"
	}
	file, err := os.Open(path)
	if err != nil {
		log.Printf("Failed to open hosts file: %v", err)
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		ip := net.ParseIP(parts[0])
		if ip == nil {
			continue
		}
		for _, name := range parts[1:] {
			name = strings.ToLower(name) + "."
			hostsMap[name] = append(hostsMap[name], ip)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Printf("Error reading hosts file: %v", err)
	}
}

func handleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
	// 1. Validate the request
	if len(r.Question) == 0 {
		dns.HandleFailed(w, r)
		return
	}

	// 2. Check hosts file
	name := strings.ToLower(r.Question[0].Name)
	if ips, ok := hostsMap[name]; ok {
		resp := new(dns.Msg)
		resp.SetReply(r)
		var answers []dns.RR
		for _, ip := range ips {
			if r.Question[0].Qtype == dns.TypeA && ip.To4() != nil {
				rr := &dns.A{
					Hdr: dns.RR_Header{Name: r.Question[0].Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 3600},
					A:   ip,
				}
				answers = append(answers, rr)
			} else if r.Question[0].Qtype == dns.TypeAAAA && ip.To4() == nil && len(ip) == net.IPv6len {
				rr := &dns.AAAA{
					Hdr:  dns.RR_Header{Name: r.Question[0].Name, Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: 3600},
					AAAA: ip,
				}
				answers = append(answers, rr)
			}
		}
		if len(answers) > 0 {
			resp.Answer = answers
			w.WriteMsg(resp)
			return
		}
	}

	key := fmt.Sprintf("%s-%d", r.Question[0].Name, r.Question[0].Qtype)

	// 3. Check Cache
	cacheMutex.RLock()
	entry, found := cache[key]
	cacheMutex.RUnlock()

	if found && time.Now().Before(entry.Expiry) {
		log.Printf("[CACHE] Hit: %s", r.Question[0].Name)
		response := entry.Msg.Copy()
		response.Id = r.Id // Sync ID with the current request
		w.WriteMsg(response)
		return
	}

	// 4. Forward via TLS if not in cache
	log.Printf("[PROXY] Querying Upstream: %s", r.Question[0].Name)
	resp, err := queryUpstreamTLS(r)
	if err != nil {
		log.Printf("[ERROR] Upstream query failed: %v", err)
		dns.HandleFailed(w, r)
		return
	}

	// 5. Store in Cache using the lowest TTL found in the answer
	if len(resp.Answer) > 0 {
		ttl := resp.Answer[0].Header().Ttl
		cacheMutex.Lock()
		cache[key] = CacheEntry{
			Msg:    resp,
			Expiry: time.Now().Add(time.Duration(ttl) * time.Second),
		}
		cacheMutex.Unlock()
	}

	w.WriteMsg(resp)
}

func queryUpstreamTLS(msg *dns.Msg) (*dns.Msg, error) {
	client := dns.Client{
		Net: "tcp-tls",
		TLSConfig: &tls.Config{
			ServerName: serverName,
		},
		Timeout: 5 * time.Second,
	}

	response, _, err := client.Exchange(msg, upstreamServer)
	return response, err
}

func main() {
	install := flag.Bool("install", false, "Install the service")
	remove := flag.Bool("remove", false, "Remove (uninstall) the service")
	start := flag.Bool("start", false, "Start the service")
	stop := flag.Bool("stop", false, "Stop the service")
	flag.Parse()

	svcConfig := &service.Config{
		Name:        "UnsinkDNS",
		DisplayName: "Unsink DNS Proxy",
		Description: "DNS-over-TLS Proxy Service",
	}

	if runtime.GOOS == "windows" {
		svcConfig.Executable = `C:\Program Files\UnsinkDNS\unsinkdns.exe`
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}

	if *install {
		if err := installService(s); err != nil {
			log.Fatal(err)
		}
		fmt.Println("Service installed successfully.")
		return
	}

	if *remove {
		if err := removeService(s); err != nil {
			log.Fatal(err)
		}
		fmt.Println("Service removed successfully.")
		return
	}

	if *start {
		err = service.Control(s, "start")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Service started successfully.")
		return
	}

	if *stop {
		err = service.Control(s, "stop")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Service stopped successfully.")
		return
	}

	// If no flags, run the service or handle other args
	if len(flag.Args()) > 0 {
		err = service.Control(s, flag.Args()[0])
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	err = s.Run()
	if err != nil {
		log.Fatal(err)
	}
}
