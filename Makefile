APP_NAME=server
MAIN=./cmd/server
SWAGGER_GEN=swag init -g ./cmd/server/main.go -o ./internal/docs --parseInternal

.PHONY: run swag watch clean

## Run server normally
run:
	go run $(MAIN)

## Generate swagger docs
swag:
	$(SWAGGER_GEN)

## Run with live reload (like nodemon)
watch:
	air -c .air.toml

## Clean swagger docs
clean:
	if exist internal\docs rmdir /s /q internal\docs
	if exist tmp rmdir /s /q tmp
