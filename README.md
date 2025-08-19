# insomnia-v5-yaml-to-bruno-converter
insomnia-v5-yaml-to-bruno-converter

# build

```bash
APP="/tmp/insomnia-v5-yaml-to-bruno-converter"; MAIN="main.go"; GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -trimpath -o "${APP}" "${MAIN}"; chmod +x "${APP}"
# APP=/tmp/app_linux;   MAIN="main.go"; GOOS=linux GOARCH=amd64   go build -ldflags="-s -w" -trimpath -o "${APP}" "${MAIN}"; chmod +x "${APP}" # linux
# APP=/tmp/app_mac;     MAIN="main.go"; GOOS=darwin GOARCH=amd64  go build -ldflags="-s -w" -trimpath -o "${APP}" "${MAIN}"; chmod +x "${APP}" # macOS
# APP=/tmp/app_win.exe; MAIN="main.go"; GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -trimpath -o "${APP}" "${MAIN}"; chmod +x "${APP}" # windows
```
