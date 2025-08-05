.PHONY: build clean dev frontend backend install run sqlc-generate

build: frontend backend

frontend:
	cd frontend && npm run build

backend:
	go build -o remote-code .

dev:
	go run .

install:
	cd frontend && npm install

sqlc-generate:
	~/go/bin/sqlc generate

clean:
	rm -rf static/*
	rm -f remote-code
	rm -f remote-code.db
	rm -f remote-code-test*.db

run: build
	./remote-code