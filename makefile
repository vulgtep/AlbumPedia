.PHONY: build run sqlc docker-build docker-run db-up db-down db-shell db-wait test-setup test-run make-test test-api

build:
	go build -o albumpedia cmd/server/main.go

run:
	go run cmd/server/main.go

sqlc:
	sqlc generate

docker-build:
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
# Se elimina db-schema del flujo, ya que Docker Compose inyectará el esquema al iniciar el volumen vacío.
test-setup: sqlc build db-down db-up db-wait

# 2) Ejecución de tests
test-run:
	go test -v ./...

# 3) Flujo completo (make-test)
make-test: test-setup
	@echo "Iniciando ejecución de tests..."
	@$(MAKE) test-run || (echo "Los tests fallaron. Limpiando contenedores..." && $(MAKE) db-down && exit 1)
	@$(MAKE) db-down
	@echo "Flujo de testing completado y entorno limpiado exitosamente."

test-api: sqlc build db-down db-up db-wait
	@sleep 3
	@$(MAKE) docker-build
	@$(MAKE) docker-run
	@echo "Esperando a que el servidor HTTP inicie..."
	@sleep 3
	@echo "Ejecutando pruebas de endpoints con Hurl..."
	@hurl --test request.hurl || (echo "Fallaron las pruebas HTTP. Limpiando entorno..." && $(MAKE) db-down && exit 1)
	@$(MAKE) db-down
	@echo "¡Pruebas de API completadas y entorno limpiado exitosamente!"