# TCP vs UDP Chat Application

This project is a Go-based terminal chat application that demonstrates the differences between TCP and UDP in real-world conditions. It includes client and server implementations for both protocols, custom session handling, real-time metrics, and testing under simulated network stress.

## Features

- Full-duplex chat using TCP and UDP
- Broadcast messaging across multiple clients
- Custom nickname registration
- Graceful disconnects (TCP) and ping-based cleanup (UDP)
- Performance metrics:
  - Latency
  - Throughput
  - Packet loss
- High-volume message stress testing (`/spam` command)

## Presentation

- [PowerPoint Presentation](https://docs.google.com/presentation/d/1rrMZR5bDRnfjTCoWR7U5MIaAHsUWP42KLefw3SsItj8/edit?usp=sharing)
- [YouTube Video (Presentation & Demo)](https://youtu.be/C_jKo7FZjUQ)

## Running the Project

### TCP

Start the TCP server:
```bash
go run tcp/main.go
```

Start a TCP client:
```bash
go run tcp/client/client.go
```

### UDP

Start the UDP server:
```bash
go run udp/main.go
```

Start a UDP client:
```bash
go run udp/client/client.go
```

### Spam Testing
In the client terminal:
```bash
/spam 1000
```

## Testing & Optimization

Tested using **Network Link Conditioner** on macOS with the following profile:
- 500ms delay
- 10% packet drop
- 1 Mbps bandwidth cap

Each client sends and receives 1000 messages. Metrics are printed on Ctrl+C exit:
- Latency (ms)
- Throughput (msg/sec)
- Packet loss (%)

## Insights

Testing locally (same machine) showed no performance differences due to loopback bypassing network simulation. Realistic testing required running clients across different machines on the same Wi-Fi network.
