# Unsunk DNS Proxy

A DNS-over-TLS proxy server that caches responses and forwards queries to Cloudflare's DNS-over-TLS server.

## Features

- DNS-over-TLS (DoT) proxy
- Caching with TTL-based expiration
- Hosts file resolution (reads /etc/hosts on Linux or C:\Windows\System32\drivers\etc\hosts on Windows, including localhost)
- Runs as a service on Windows and Linux

## Building

### For Windows
```bash
go build -o unsink-dns.exe
```

### For Linux
```bash
GOOS=linux GOARCH=amd64 go build -o unsink-dns
```

## Installation as Service

The application uses the `github.com/kardianos/service` package to run as a system service.

### Windows

1. Build the executable: `go build -o unsink-dns.exe`
2. Install the service: `unsink-dns.exe --install`
3. Start the service: `unsink-dns.exe --start`
4. Stop the service: `unsink-dns.exe --stop`
5. Remove the service: `unsink-dns.exe --remove`

Note: You may need to run the commands as Administrator.

### Linux

1. Build the executable: `GOOS=linux GOARCH=amd64 go build -o unsink-dns`
2. Copy the binary to `/usr/local/bin/` or a suitable location: `sudo cp unsink-dns /usr/local/bin/`
3. Install the service: `sudo unsink-dns --install`
4. Start the service: `sudo unsink-dns --start`
5. Stop the service: `sudo unsink-dns --stop`
6. Remove the service: `sudo unsink-dns --remove`

Note: On Linux, you may need to configure systemd or other init systems if the service package doesn't handle it automatically. The `kardianos/service` package supports systemd on Linux.

## Running Manually

To run the application manually (not as a service):

```bash
./unsink-dns
```

Or on Windows: `unsink-dns.exe`

The server will listen on UDP port 53.

## Configuration

Currently, the upstream server is hardcoded to Cloudflare (1.1.1.1:853). To change it, modify the `upstreamServer` and `serverName` variables in `main.go`.

## Troubleshooting

- Ensure port 53 is not already in use.
- On Linux, you may need to run with elevated privileges or configure capabilities.
- Check logs for errors.

## Dependencies

- `github.com/miekg/dns`
- `github.com/kardianos/service`