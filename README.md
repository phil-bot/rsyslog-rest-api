# rsyslog REST API

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.21+-00ADD8.svg)](https://go.dev/)
[![Release](https://img.shields.io/github/v/release/phil-bot/rsyslog-rest-api)](https://github.com/phil-bot/rsyslog-rest-api/releases)

High-performance REST API for rsyslog/MySQL written in Go.

## 📖 About

A modern REST API server that queries rsyslog data from a MySQL/MariaDB database and makes it accessible via HTTP/JSON. Perfect for monitoring dashboards, log analysis, and system integration.

### 🌟 Features

- 🚀 **High Performance** - Compiled in Go for maximum speed
- 🔍 **Advanced Filtering** - Multi-value filters for complex queries
- 📊 **All Fields** - Access to all 25+ SystemEvents columns
- 🔐 **Secure** - API key authentication, SSL/TLS support
- 🐳 **Docker Ready** - Complete test environment with live data
- 📝 **REST API** - Clean JSON responses
- 🎯 **RFC-5424 Compliant** - Proper syslog severity and facility labels

## 🚀 Quick Start

### Binary Installation (Recommended)

```bash
# Download latest release
wget https://github.com/phil-bot/rsyslog-rest-api/releases/latest/download/rsyslog-rest-api-linux-amd64

# Install
chmod +x rsyslog-rest-api-linux-amd64
sudo mv rsyslog-rest-api-linux-amd64 /usr/local/bin/rsyslog-rest-api

# Create configuration
cat > .env << EOF
API_KEY=$(openssl rand -hex 32)
DB_HOST=localhost
DB_NAME=Syslog
DB_USER=rsyslog
DB_PASS=your-password
EOF

# Run
rsyslog-rest-api
```

**Test the API:**
```bash
curl http://localhost:8000/health
curl -H "X-API-Key: YOUR_KEY" "http://localhost:8000/logs?limit=5"
```

### Docker Test Environment

Perfect for testing with live generated logs:

```bash
# Build binary
make build-static

# Start container
cd docker && docker-compose up -d

# Test
curl "http://localhost:8000/logs?limit=5"
```

→ [Full Installation Guide](docs/installation.md)

## 📚 Documentation

### Getting Started
- [**Installation Guide**](docs/installation.md) - All installation methods
- [**Configuration**](docs/configuration.md) - Complete configuration
- [**Quick Examples**](docs/examples.md) - Practical examples

### API & Usage
- [**API Reference**](docs/api-reference.md) - All endpoints and parameters
- [**Troubleshooting**](docs/troubleshooting.md) - Troubleshooting and FAQ

### Administration
- [**Deployment**](docs/deployment.md) - Production setup
- [**Security**](docs/security.md) - Security best practices

### Development
- [**Docker Testing**](docs/docker.md) - Test environment
- [**Development**](docs/development.md) - Architecture and contributing

→ [**Full Documentation**](docs/index.md)

## 💡 Examples

### Retrieve logs with filters

```bash
# All errors from the last hour
curl -H "X-API-Key: YOUR_KEY" \
  "http://localhost:8000/logs?Priority=3&start_date=2025-02-09T09:00:00Z"

# Logs from multiple hosts (Multi-value!)
curl -H "X-API-Key: YOUR_KEY" \
  "http://localhost:8000/logs?FromHost=web01&FromHost=web02&FromHost=db01"

# Combined filters
curl -H "X-API-Key: YOUR_KEY" \
  "http://localhost:8000/logs?FromHost=web01&Priority=3&Priority=4&limit=20"
```

### Query metadata

```bash
# All available hosts
curl -H "X-API-Key: YOUR_KEY" "http://localhost:8000/meta/FromHost"

# All priorities with labels
curl -H "X-API-Key: YOUR_KEY" "http://localhost:8000/meta/Priority"

# Hosts that had errors
curl -H "X-API-Key: YOUR_KEY" \
  "http://localhost:8000/meta/FromHost?Priority=3&Priority=4"
```

→ [More Examples](docs/examples.md)

## 🆕 What's New in v0.2.2?

- ✅ **Multi-Value Filters** - Multiple values per parameter
- ✅ **Extended Columns** - All 25+ SystemEvents fields
- ✅ **Live Log Generator** - Realistic test logs (Docker)
- ✅ **Meta Endpoint** - Now also filters with multi-value

→ [Changelog](docs/changelog.md)

## 🗺️ Roadmap

### v0.3.0 (Planned)
- Negation filters (`exclude`, `not`)
- Advanced filter combinations
- Complex query support

### v0.4.0 (Planned)
- Statistics endpoint (`/stats`)
- Aggregations
- Timeline/Histogram

→ [GitHub Issues](https://github.com/phil-bot/rsyslog-rest-api/issues)

## 🤝 Support & Community

- **Issues:** [GitHub Issues](https://github.com/phil-bot/rsyslog-rest-api/issues)
- **Discussions:** [GitHub Discussions](https://github.com/phil-bot/rsyslog-rest-api/discussions)
- **Documentation:** [docs/](docs/index.md)

## 🙏 Contributing

Contributions are welcome! Please read the [Contributing Guidelines](docs/development.md#contributing).

1. Fork the repository
2. Create feature branch
3. Make changes & add tests
4. Submit pull request

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

## ✨ Credits

Created with ❤️ for the syslog community.

**Built with:**
- [Go](https://go.dev/)
- [go-sql-driver/mysql](https://github.com/go-sql-driver/mysql)
- [rsyslog](https://www.rsyslog.com/)
- [MariaDB](https://mariadb.org/)

---

⭐ **Star this project** if it helps you!
