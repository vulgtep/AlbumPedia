.PHONY: build run sqlc templ start docker-build docker-run db-up db-down db-shell db-wait test-setup test-run make-test

sqlc:
	sqlc generate

templ:
	templ generate

build: sqlc templ
	go build -o albumpedia cmd/server/main.go

run: sqlc templ
	go run cmd/server/main.go

start: db-up db-wait run

docker-build: sqlc templ
	docker compose build

docker-run:
	docker compose up -d app

db-up:
	docker compose up -d db

db-down:
	docker compose down -v

db-shell:
	docker compose exec db psql -U postgres -d albumpedia

db-wait:
	@echo "Esperando a que la base de datos esté lista..."
	@for i in 1 2 3 4 5 6 7 8 9 10; do \
		docker compose exec db pg_isready -U postgres > /dev/null 2>&1 && echo "¡DB Lista!" && break || (echo "Esperando..." && sleep 2); \
	done

# 1) Tareas previas
test-setup: sqlc templ build db-down db-up db-wait

# 2) Ejecución de tests
test-run:
	go test -v ./...

# 3) Flujo completo (make-test)
make-test: test-setup
	@echo "Iniciando ejecución de tests..."
	@$(MAKE) test-run || (echo "Los tests fallaron. Limpiando contenedores..." && $(MAKE) db-down && exit 1)
	@$(MAKE) db-down
	@echo "Flujo de testing completado y entorno limpiado exitosamente."
