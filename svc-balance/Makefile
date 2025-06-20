sqlc:
	cd store && sqlc generate

test:
	go test ./service ./activity ./util/failure -v

.PHONY: sqlc test