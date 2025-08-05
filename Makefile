.PHONY: build clean dev frontend backend install run sqlc-generate

build: frontend backend

frontend:
	cd frontend && npm run build

backend:
	go build -o web-terminal .

dev:
	go run .

install:
	cd frontend && npm install

sqlc-generate:
	~/go/bin/sqlc generate

clean:
	rm -rf static/*
	rm -f web-terminal
	rm -f web-terminal.db

run: build
	./web-terminal