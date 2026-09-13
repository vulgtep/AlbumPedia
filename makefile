.PHONY: build run sqlc docker-build docker-run db-up db-down db-schema db-shell db-wait test-setup test-run make-test

build:
	go build -o albumpedia main.go

run:
	go run main.go

sqlc:
	sqlc generate

docker-build:
	docker build -t albumpedia .

docker-run:
	docker run -p 8080:8080 albumpedia

db-up:
	-docker stop albumpedia-db
	-docker rm albumpedia-db
	docker run --name albumpedia-db -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=albumpedia -p 5432:5432 -d postgres:16-alpine

# Modificado para ignorar errores si el contenedor no existe al intentar borrarlo
db-down:
	-docker stop albumpedia-db 2>/dev/null || true
	-docker rm -v albumpedia-db 2>/dev/null || true

# Nota: Ajusté la ruta a DB/schema/schema.sql basándome en la estructura de tu ZIP original
db-schema:
	until cat DB/schema/schema.sql | docker exec -i albumpedia-db psql -U postgres -d albumpedia; do sleep 1; done

db-shell:
	docker exec -it albumpedia-db psql -U postgres -d albumpedia

# --- NUEVAS TAREAS PARA EL FLUJO DE TESTING ---

db-wait:
	@echo "Esperando a que la base de datos esté lista..."
	@for i in 1 2 3 4 5 6 7 8 9 10; do \
		docker exec albumpedia-db pg_isready -U postgres > /dev/null 2>&1 && echo "¡DB Lista!" && break || (echo "Esperando..." && sleep 2); \
	done

# 1) Tareas previas
test-setup: sqlc build db-down db-up db-wait db-schema

# 2) Ejecución de tests
test-run:
	go test -v ./...

# 3) Flujo completo (make-test)
make-test: test-setup
	@echo "Iniciando ejecución de tests..."
	@$(MAKE) test-run || (echo "Los tests fallaron. Limpiando contenedores..." && $(MAKE) db-down && exit 1)
	@$(MAKE) db-down
	@echo "Flujo de testing completado y entorno limpiado exitosamente."