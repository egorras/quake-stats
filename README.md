# Quake Stats Collector

A Go application that connects to a Quake server via ZeroMQ to collect and process game events.

## Quick Start

### Linux/macOS
```bash
cd src/collector
go build
./collector
```

### Windows
See [Windows Setup](windows-setup.md) for Docker-based testing.

## Configuration

```yaml
# config.yaml
zmq_endpoint: tcp://89.168.29.137:27960
batch_size: 10
flush_interval_sec: 1
verbose_logging: true
```

## Deployment

GitHub Actions automatically deploys to a VM as a systemd service. See `.github/workflows/deploy.yml`. 