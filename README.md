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

---

# Interfaz de Usuario Cliente (TP4)

En esta etapa se desarrolló el frontend del proyecto utilizando **HTML5, CSS3 (Flexbox y CSS Grid) y JavaScript nativo** (`static/index.html`, `static/styles.css` y `static/app.js`) para consumir de forma asíncrona la API REST del servidor Go:

* **Listado dinámico (`GET /albums`):** Al cargar la página (`DOMContentLoaded`), el cliente realiza una petición `fetch` para obtener los álbumes en formato JSON y genera las tarjetas dinámicamente en el DOM, incluyendo un visor interactivo con transición animada entre portada y contraportada.
* **Alta de álbumes (`POST /albums`):** El formulario captura el evento `submit`, previene la recarga por defecto (`event.preventDefault()`), valida los campos obligatorios (`titulo`, `banda` y `genero`) y envía el payload JSON a la API, refrescando el listado al finalizar.
* **Borrado con Delegación de Eventos (`DELETE /albums/{id}`):** Se implementó un único listener sobre el contenedor `#albums-container` para interceptar los clics en los botones de borrado generados dinámicamente, eliminando el registro en el servidor y actualizando la vista.

---

# Server-Side Rendering con `templ` (TP5)

## Cambios de Arquitectura

En el TP5 se reemplazó la arquitectura de cliente separado (API JSON + Fetch en JS) por una aplicación **Server-Side Rendered (SSR)** integrada directamente en el servidor Go utilizando **[`templ`](https://templ.guide/)**.

### 1. Componentes de Vista (`views/`)
La interfaz fue dividida en componentes `.templ` fuertemente tipados que se compilan a código Go nativo (`views/*_templ.go`):

* **`views/layout.templ`:** Define el componente `Layout(title string)` con la estructura base del documento HTML (`<html>`, `<head>`, `<body>`), integrando **Pico.css** junto con `/static/styles.css`, y el componente `HomePage(albums []sqlc.Album)` que compone el formulario y la lista.
* **`views/entity_form.templ`:** Define `EntityForm()`, que renderiza el formulario clásico `<form method="POST" action="/albums">` con los campos obligatorios (`titulo`, `banda`, `genero`) y desplegable de campos opcionales.
* **`views/entity_list.templ`:** Define `EntityList(albums []sqlc.Album)`, que recibe el slice `[]sqlc.Album` generado por `sqlc` y renderiza en el servidor el catálogo de tarjetas, el visor de portada/contraportada y el formulario de borrado.

### 2. Refactorización del Servidor y Patrón Post/Redirect/Get (PRG)
Se actualizaron `cmd/server/main.go` y `pkg/handlers/albums.go` para gestionar el ciclo SSR clásico:

| Método | Ruta | Handler | Descripción |
| :--- | :--- | :--- | :--- |
| `GET` | `/` | `HomeHandler` | Obtiene todos los álbumes con `Queries.ListAlbums` y renderiza `views.HomePage(albums)` en el `http.ResponseWriter`. |
| `POST` | `/albums` | `CreateAlbumHandler` | Parsea los datos del formulario (`r.ParseForm()`), inserta el nuevo álbum con `Queries.CreateAlbum` y redirige a `/` mediante `http.Redirect` (`HTTP 303 See Other`), implementando el patrón **Post/Redirect/Get (PRG)**. |
| `POST` | `/albums/delete` | `DeleteAlbumHandler` | Elimina un álbum por su `id` usando `Queries.DeleteAlbum` y redirige a `/` aplicando el patrón **PRG**. |

## Instrucciones de Generación y Ejecución (TP5)

**Requisitos previos:** Docker, Go, `sqlc`, `templ` y `make`.

Para instalar el CLI de `templ` (si no está instalado):
```bash
go install github.com/a-h/templ/cmd/templ@v0.3.857
```

### Forma automática (con un solo comando)

Se agregó el comando `make start` en el `makefile`, el cual levanta PostgreSQL en Docker, espera a que la base de datos esté lista, genera el código de `sqlc` y `templ`, y ejecuta el servidor Go:

```bash
make start
```

### Forma paso a paso

Si preferís ejecutar cada etapa manualmente:

```bash
make db-up      # 1. Levanta el contenedor de PostgreSQL en segundo plano
make db-wait    # 2. Espera a que la base de datos acepte conexiones
make sqlc       # 3. Genera el código Go de acceso a datos (DB/sqlc)
make templ      # 4. Compila los componentes .templ a código Go (views/*_templ.go)
make run        # 5. Ejecuta el servidor en http://localhost:8080
```

Una vez iniciado el servidor, abrí tu navegador en **http://localhost:8080** para visualizar el listado de álbumes y crear nuevos registros desde el formulario con recarga SSR completa.

Para detener los contenedores y limpiar los volúmenes al finalizar:

```bash
make db-down
```
