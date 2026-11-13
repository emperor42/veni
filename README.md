# VENI

VENI is two small, independent pieces:

1. `veni.js` is a browser helper for discovering and registering custom-element tags already present in a page.
2. `main.go` is an optional visual web-crawler demo. It is useful for trying the browser UI locally; it is not the component library and it is not a production web-crawler service.

## Overview

`veni` is a Go backend service that provides APIs for custom element discovery and management, serving as the foundation for the VENI browser utility.

## Features

- **Custom Element Discovery**: API to discover and register custom elements
- **Browser Integration**: RESTful APIs for browser component management
- **Health Monitoring**: Built-in health check endpoint
- **Rate Limiting**: Protection against abuse
- **Security First**: Implements security best practices

## API Endpoints

### Discovery API
- **GET /api/discover** - Discover custom elements (rate limited)
- **GET /api/elements** - List all available elements

### System Status
- **GET /health** - Health check endpoint

## Technology Stack

- **Backend**: Go
- **API**: RESTful JSON
- **Security**: HTTP security headers, rate limiting
- **Deployment**: Docker ready

## Running

### Prerequisites

```bash
go 1.21+
```

### Building and Running

```bash
# Build the application
go build

# Run the application
./veni

# Or using Docker
docker build -t veni .
docker run -p 8082:8082 veni
```

### Testing

```bash
# Run tests
go test -v ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Usage

### Local Development

```bash
# Start the server
./veni

# Discover custom elements
curl http://localhost:8082/api/discover

# List all elements
curl http://localhost:8082/api/elements

# Check system health
curl http://localhost:8082/health
```

### API Examples

#### Discover Custom Elements
```bash
curl -X GET http://localhost:8082/api/discover
```

#### List Elements
```bash
curl -X GET http://localhost:8082/api/elements
```

#### Health Check
```bash
curl -X GET http://localhost:8082/health
```

## Deployment

### Docker Deployment

```Dockerfile
FROM golang:1.21-alpine

WORKDIR /app

COPY go.mod go.sum ./
COPY . .

RUN go build -o veni main.go

EXPOSE 8082

CMD ["./veni"]
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: veni
  labels:
    app: veni
spec:
  replicas: 2
  selector:
    matchLabels:
      app: veni
  template:
    metadata:
      labels:
        app: veni
    spec:
      containers:
      - name: veni
        image: veni:latest
        ports:
        - containerPort: 8082
        env:
        - name: NODE_ENV
          value: "production"
```

## Development

### Project Structure

```
.
├── main.go          # Main application
├── main_test.go     # Tests
├── go.mod          # Module configuration
└── README.md       # Documentation
```

### Features Implemented

- **Custom Element Discovery**: API for discovering browser custom elements
- **Rate Limiting**: Protection against API abuse
- **Security First**: Security headers and validation
- **API-First**: RESTful API design
- **Developer Experience**: Comprehensive testing

## License

MIT License

## Future Enhancements

- [ ] Add element categorization
- [ ] Implement element metadata
- [ ] Add batch discovery APIs
- [ ] Implement element registration
- [ ] Add performance monitoring
- [ ] Integrate with browser extension

## Contributing

1. Fork the repository
2. Create a feature branch
3. Implement your changes
4. Add tests
5. Submit a pull request

## Contact

For questions or suggestions, please open an issue or contact the maintainer.