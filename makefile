APP_NAME= ws-event
build:
	go build -o bin/$(APP_NAME) cmd/main.go
test:
	k6 run websocket-bench-long.js
	k6 run websocket-bench-short.js
	k6 run websocket-bench-short.js
