# Albumpedia - TP1

## Instrucciones 

### 1. Clonar el repositorio
Abre tu terminal y ejecuta el siguiente comando para clonar el proyecto directamente en la rama correspondiente:

```bash
git clone -b tp1 <https://github.com/vulgtep/AlbumPedia.git>
```

### 2. Ingresar a la carpeta
```bash
cd Albumpedia
```

### 3. Ejecutar el servidor
```bash
go run main.go
```

### 4. Visualizar la página
Abri el navegador e ingresa a http://localhost:8080

# Ejecución de Tests (TP2)

El proyecto incluye pruebas automatizadas creadas con el paquete nativo `testing` de Go (`queries_test.go`). Todo el ciclo de vida de las pruebas está orquestado con `make`.

Para ejecutar las pruebas, asegúrate de tener Docker abierto y corre el siguiente comando en la raíz del proyecto:

```bash
make make-test
```

**Este comando automatiza el siguiente flujo:**
1. **Tareas Previas:** Genera el código `sqlc`, compila el proyecto, borra contenedores/volúmenes antiguos, levanta un nuevo PostgreSQL en Docker, espera su inicio e inyecta el esquema de tablas.
2. **Ejecución:** Corre los tests de Go (`go test -v ./...`).
3. **Limpieza Posterior:** Borra el contenedor de base de datos y sus volúmenes, independientemente del resultado de los tests.