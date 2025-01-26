# Go Key-Value Store

Simple REST key-value store with JSON persistence

## Usage

```bash
# Start server
go run main.go

# Add key-value
curl "localhost:8080/PUT?key=hello&value=world"

# Get value
curl "localhost:8080/GET?key=hello"

# Delete key
curl "localhost:8080/DELETE?key=hello"
