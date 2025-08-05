.PHONY: build clean dev frontend backend install run

build: frontend backend

frontend:
	cd frontend && npm run build

backend:
	go build -o web-terminal main.go

dev:
	go run main.go

install:
	cd frontend && npm install

clean:
	rm -rf static/*
	rm -f web-terminal

run: build
	./web-terminal