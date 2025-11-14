echo "Starting API..."
go run ./cmd/api/main.go &

echo "Starting Worker..."
go run ./cmd/worker/main.go &

echo "All running..."
