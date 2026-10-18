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

### Estructura de la Tabla: `albums`

La persistencia central del proyecto se apoya en la tabla `albums` dentro de PostgreSQL, la cual define las propiedades esenciales de cada disco.

| Columna | Tipo de Dato | Restricciones | Descripción |
| :--- | :--- | :--- | :--- |
| `id` | SERIAL | PRIMARY KEY | Identificador único autoincremental del álbum. |
| `titulo` | VARCHAR(150) | NOT NULL | Título o nombre del álbum. |
| `genero` | VARCHAR(100) | NOT NULL | Género musical principal del álbum. |
| `banda` | VARCHAR(150) | NOT NULL | Nombre de la banda o artista creador. |
| `productor` | VARCHAR(150) | Permite NULL | Nombre del productor o productores a cargo. |
| `lista_artistas` | VARCHAR(150) | Permite NULL | Listado de integrantes o artistas invitados. |
| `portada` | VARCHAR(255) | Permite NULL | Enlace o ruta de archivo para la imagen frontal. |
| `contraportada` | VARCHAR(255) | Permite NULL | Enlace o ruta de archivo para la imagen trasera. |
| `duracion_minutos` | INTEGER | Permite NULL | Duración total del álbum expresada en minutos. |
| `fecha_salida` | DATE | Permite NULL | Fecha oficial de publicación o lanzamiento. |
| `link_escucha` | VARCHAR(500) | Permite NULL | URL directa para reproducir el álbum (Spotify, YouTube, etc.). |
| `observacion` | VARCHAR(1000)| Permite NULL | Campo de texto amplio para reseñas u opiniones personales. |

# Ejecución de Tests (TP2)

El proyecto incluye pruebas automatizadas creadas con el paquete nativo `testing` de Go (`queries_test.go`).

Para ejecutar las pruebas, corre el siguiente comando en la raíz del proyecto:

```bash
make make-test
```

**Este comando automatiza el siguiente flujo:**
1. **Tareas Previas:** Genera el código `sqlc`, compila el proyecto, borra contenedores/volúmenes antiguos, levanta un nuevo PostgreSQL en Docker, espera su inicio e inyecta el esquema de tablas.
2. **Ejecución:** Corre los tests de Go (`go test -v ./...`).
3. **Limpieza Posterior:** Borra el contenedor de base de datos y sus volúmenes, independientemente del resultado de los tests.

---

# API REST, Pruebas HTTP (TP3)

## Modificaciones de Infraestructura y Orquestación

Incorporamos el archivo `docker-compose.yml` para centralizar la definición y ejecución de múltiples contenedores mediante un archivo YAML. Debido a esto, se modificó el `makefile` para delegar la orquestación de la infraestructura a Docker Compose:

* **`docker-build` y `docker-run`:** Sustituidos por `docker compose build` y `docker compose up -d app` para construir la imagen y levantar el servicio de la API en segundo plano utilizando la configuración del archivo YAML.
* **`db-up`:** Reemplazado por `docker compose up -d db`. Ya no es necesario inyectar variables de entorno explícitas (como `POSTGRES_PASSWORD`) en el `makefile`, ya que estas se definen directamente en `docker-compose.yml`.
* **`db-down`:** Modificado a `docker compose down -v` para detener los contenedores y eliminar los volúmenes de datos asociados, garantizando un entorno limpio en cada ejecución de la suite de testing.
* **`db-schema` (Eliminado):** La tarea manual de inyección del esquema fue removida del `makefile` y de la cadena de ejecución de `test-setup`. Al emplear volúmenes en `docker-compose.yml` para mapear el archivo `schema.sql` al directorio de inicialización (`/docker-entrypoint-initdb.d/init.sql`), PostgreSQL aplica la estructura automáticamente al crearse la base de datos.
* **`db-shell` y `db-wait`:** Actualizados para utilizar `docker compose exec db` en lugar de `docker exec albumpedia-db`, interactuando con el contenedor a través de la capa de servicio definida por Compose en lugar de apuntar al nombre estático del contenedor.

## Instrucciones para levantar el servidor, conexión con servidor y testeo

**Requisitos previos:** Docker, Go, sqlc, make y hurl.

### Forma automática

Para realizar todo el circuito de testeo (desde la generación de sqlc hasta la prueba con hurl) con un solo comando:

```bash
make test-api
```

Este comando ejecuta todo el circuito y, al finalizar, elimina los volúmenes y contenedores.

### Comandos de forma manual

Para levantar el servidor y conectar con la base de datos se deben ejecutar los siguientes comandos del makefile en este orden:

```bash
make sqlc
make build
make db-down
make db-up
make db-wait
make run
```

Con estos comandos el servidor y la base de datos quedarán conectados y funcionando en el puerto 8080.

## Testeo con hurl

Para realizar los tests con hurl, primero se deben ejecutar los pasos anteriores y dejar esa terminal corriendo. Luego, en una nueva consola, ejecutar:

```bash
hurl --test request.hurl
```

Una vez finalizadas las pruebas, para detener los contenedores y limpiar los volúmenes ejecutar:

```bash
make db-down
```

## Documentación de la API

La API está documentada con OpenAPI 3.0 en [`docs/openapi.yaml`](docs/openapi.yaml).

Para verla de forma interactiva:

1. Entrá a https://editor.swagger.io/
2. Pegá el contenido de `docs/openapi.yaml`.

### Endpoints

| Método | Ruta         | Descripción         |
|--------|--------------|---------------------|
| GET    | /albums      | Listar álbumes      |
| POST   | /albums      | Crear un álbum      |
| GET    | /albums/{id} | Obtener un álbum    |
| PUT    | /albums/{id} | Actualizar un álbum |
| DELETE | /albums/{id} | Eliminar un álbum   |