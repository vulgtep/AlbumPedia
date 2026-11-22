# AlbumPedia — Trabajo Práctico Integrador (TP6: Interactividad Tipo SPA con HTMX y `templ`)

Aplicación web en **Go** para gestionar y catalogar álbumes de música con persistencia en **PostgreSQL** (`sqlc`), orquestación mediante **Docker Compose**, renderizado en el servidor (**SSR**) con **[`templ`](https://templ.guide/)** e interactividad dinámica sin recargas de página utilizando **[HTMX](https://htmx.org/)**.

---

## Inicio Rápido e Instalación (TP6)

### Requisitos Previos
* **Go** (1.24+)
* **Docker** y **Docker Compose**
* **`sqlc`** (`go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`)
* **`templ`** (`go install github.com/a-h/templ/cmd/templ@v0.3.857`)
* **`make`**

### 1. Clonar el repositorio en la rama del TP6
```bash
git clone -b tp6 https://github.com/vulgtep/AlbumPedia.git
cd AlbumPedia
```

### 2. Configurar las variables de entorno (`.env`)
Las credenciales y parámetros de conexión **no están hardcodeados** en `docker-compose.yml` ni en el código fuente. Para configurar tu entorno local, copiá el archivo de plantilla `.env.example` a `.env` y definí tus credenciales:

```bash
cp .env.example .env
```

Variables utilizadas en `.env`:
| Variable | Descripción | Ejemplo (Desarrollo Local) |
| :--- | :--- | :--- |
| `POSTGRES_USER` | Usuario del motor PostgreSQL en Docker | `postgres` |
| `POSTGRES_PASSWORD` | Contraseña del motor PostgreSQL | `tu_password_seguro` |
| `POSTGRES_DB` | Nombre de la base de datos en PostgreSQL | `albumpedia` |
| `DB_HOST` | Host de conexión para el servidor Go | `localhost` (o `db` en contenedor) |
| `DB_PORT` | Puerto expuesto por PostgreSQL | `5432` |
| `DB_USER` | Usuario de conexión de la aplicación Go | `postgres` |
| `DB_PASSWORD` | Contraseña de conexión de la aplicación Go | `tu_password_seguro` |
| `DB_NAME` | Nombre de la base de datos a utilizar | `albumpedia` |
| `APP_PORT` | Puerto HTTP del servidor web | `8080` |

> **Nota de seguridad:** El archivo `.env` está excluido del control de versiones mediante `.gitignore` para evitar exponer credenciales sensibles en el repositorio.

### 3. Ejecución automática (con un solo comando)
El comando `make start` carga las variables de `.env`, levanta el contenedor de PostgreSQL, espera a que acepte conexiones, genera el código de `sqlc` y `templ`, e inicia el servidor:

```bash
make start
```

### 4. Ejecución paso a paso (manual)
Si preferís ejecutar cada etapa por separado:

```bash
make db-up      # 1. Levanta el contenedor de PostgreSQL en segundo plano
make db-wait    # 2. Espera a que la base de datos acepte conexiones
make sqlc       # 3. Genera el código Go de acceso a datos (DB/sqlc)
make templ      # 4. Compila los componentes .templ a código Go (views/*_templ.go)
make run        # 5. Ejecuta el servidor en http://localhost:8080
```

### 5. Visualizar la aplicación
Abrí el navegador e ingresá a **http://localhost:8080**. Podrás crear y eliminar álbumes viendo cómo la lista se actualiza instantáneamente sin recargar la página completa.

Para detener los contenedores y eliminar los volúmenes al finalizar:
```bash
make db-down
```

---

## Arquitectura Reactiva con HTMX y `templ` (TP6)

En el TP6 se evolucionó la aplicación renderizada en el servidor (`templ` + Go) del TP5 para que se comporte como una **Single Page Application (SPA)** utilizando **[HTMX](https://htmx.org/)**, reemplazando las redirecciones de página completa por actualizaciones parciales de fragmentos HTML en el DOM.

### 1. Integración de HTMX en el Layout (`views/layout.templ`)
* Se añadió el script de HTMX (`<script src="https://unpkg.com/htmx.org@2.0.4"></script>`) dentro del `<head>` del componente `Layout`.
* El listener de transición animada entre portada y contraportada se delegó a nivel `document` en el layout base para que siga funcionando de manera transparente tras cualquier actualización parcial del contenedor `#albums-list`.

### 2. Creación sin Recarga de Página (`POST /albums`)
* **`views/entity_form.templ`:** El `<form id="album-form">` utiliza atributos declarativos de HTMX:
  * `hx-post="/albums"`: envía los datos del formulario en segundo plano mediante una petición AJAX `POST`.
  * `hx-target="#albums-list"`: apunta al ID del contenedor principal del componente `EntityList` (`<section id="albums-list">`).
  * `hx-swap="outerHTML"`: reemplaza el contenedor de la lista con el fragmento HTML actualizado que devuelve el servidor.
  * `hx-on::after-request="if(event.detail.successful) this.reset()"`: limpia automáticamente los campos del formulario tras una creación exitosa.
* **`pkg/handlers/albums.go` (`CreateAlbumHandler`):**
  * Se eliminó la redirección `http.Redirect`.
  * Luego de insertar el nuevo álbum en PostgreSQL con `Queries.CreateAlbum`, se vuelve a consultar el listado completo mediante `Queries.ListAlbums` y se renderiza y devuelve **únicamente** el componente parcial `views.EntityList(albums)`.

### 3. Borrado sin Recarga de Página (`DELETE /albums/{id}`)
* **`views/entity_list.templ`:** Cada tarjeta de álbum (`<article class="album-card">`) cuenta con un botón de eliminación configurado con atributos HTMX:
  * `hx-delete="/albums/{id}"`: envía una petición HTTP `DELETE` a la ruta del álbum específico.
  * `hx-target="closest article"`: indica a HTMX que el elemento objetivo a actualizar/eliminar del DOM es la tarjeta `<article>` que contiene al álbum.
  * `hx-swap="outerHTML"`: reemplaza la tarjeta `<article>` completa con la respuesta vacía del servidor, quitándola de la vista sin recargar la página.
  * `hx-confirm`: solicita confirmación al usuario antes de ejecutar el borrado.
* **`pkg/handlers/albums.go` (`DeleteAlbumHandler`) y `cmd/server/main.go`:**
  * Se registró el handler `DELETE /albums/{id}` que extrae el ID de la URL, invoca `Queries.DeleteAlbum` de `sqlc` y devuelve una respuesta vacía con código `200 OK` (`w.WriteHeader(http.StatusOK)`).

### Resumen de Rutas HTTP (TP6)

| Método | Ruta | Handler | Respuesta del Servidor / Efecto en HTMX |
| :--- | :--- | :--- | :--- |
| `GET` | `/` | `HomeHandler` | Renderiza la página inicial completa (`views.HomePage(albums)`). |
| `POST` | `/albums` | `CreateAlbumHandler` | Crea el álbum y devuelve el fragmento parcial `views.EntityList(albums)` que reemplaza `#albums-list`. |
| `DELETE` | `/albums/{id}` | `DeleteAlbumHandler` | Elimina el álbum en BD y devuelve `200 OK` vacío, eliminando el `closest article` del DOM. |

---

## Modelo de Datos: Tabla `albums`

La persistencia del proyecto se apoya en la tabla `albums` dentro de PostgreSQL (`DB/schema/schema.sql`):

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

---

## Ejecución de Tests de Persistencia

El proyecto incluye pruebas automatizadas con el paquete `testing` de Go (`DB/queries_test.go`). Para ejecutarlas:

```bash
make make-test
```

Este comando automatiza la generación de `sqlc` y `templ`, la compilación, el levantado limpio de PostgreSQL en Docker, la ejecución de `go test -v ./...` y la limpieza final de contenedores y volúmenes.

---

## Historial de Etapas Anteriores (TP1 – TP5)

* **TP1 y TP2 (Modelo de Datos y Persistencia con `sqlc`):** Diseño del esquema relacional en PostgreSQL (`DB/schema/schema.sql`), consultas tipadas con `sqlc` (`DB/queries/query.sql`) y suite de pruebas unitarias de base de datos (`DB/queries_test.go`).
* **TP3 (Docker Compose y API REST):** Orquestación de contenedores con `docker-compose.yml`, documentación OpenAPI 3.0 (`Openapi.yaml`) y pruebas HTTP automatizadas con `hurl` (`request.hurl`).
* **TP4 (Frontend Cliente HTML/CSS/JS):** Interfaz estática con CSS Grid/Flexbox (`static/styles.css`) y visor interactivo con transición animada entre portada y contraportada.
* **TP5 (Server-Side Rendering con `templ`):** Refactorización de las vistas a componentes `.templ` (`views/layout.templ`, `views/entity_form.templ`, `views/entity_list.templ`) renderizados directamente por el servidor Go con el patrón Post/Redirect/Get (PRG).
