# Add MinGW-w64 to PATH
$env:PATH = "C:\msys64\mingw64\bin;$env:PATH"

# Configure Go build environment
$env:GOOS = "windows"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"  # Disable CGO since we're not using taglib for now

Write-Host "Note: Building without CGO. MP3 tagging functionality will be disabled."
Write-Host "Installing Go dependencies..."
go mod tidy

Write-Host "Building server..."

# Create bin directory if it doesn't exist
if (-not (Test-Path "bin")) {
    New-Item -ItemType Directory -Path "bin"
}

# Build the server with build tags to exclude taglib
go build -tags "notag" -o bin/beatportdl-server.exe ./cmd/server 