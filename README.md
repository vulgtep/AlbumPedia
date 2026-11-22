# TP6: Interactividad Tipo SPA con HTMX y `templ`

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
| `SESSION_SECRET` | Clave secreta para firmar las cookies de sesión | `tu_secreto_aleatorio` |

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
| `GET` | `/` | `HomeHandler` | Renderiza el perfil del usuario conectado con el formulario y sus álbumes (`views.HomePage`). |
| `GET / POST` | `/login` | `LoginHandler` | Pantalla de inicio de sesión y validación de credenciales. |
| `GET / POST` | `/register` | `RegisterHandler` | Pantalla de registro para crear una cuenta nueva. |
| `POST` | `/logout` | `LogoutHandler` | Cierra la sesión actual. |
| `GET` | `/users/search` | `SearchUsersHandler` | Busca usuarios por nombre y devuelve la lista desplegable en vivo con HTMX. |
| `GET` | `/u/{username}` | `UserProfileHandler` | Muestra el perfil y los álbumes de otro usuario en modo solo lectura. |
| `POST` | `/albums` | `CreateAlbumHandler` | Crea el álbum en la cuenta actual y devuelve el fragmento `views.EntityList` que actualiza `#albums-list`. |
| `GET` | `/albums/{id}/edit` | `DeleteAlbumHandler` | Abre la ventana emergente flotante con los datos del álbum para editarlo (solo si sos el dueño). |
| `PUT` | `/albums/{id}` | `DeleteAlbumHandler` | Guarda los cambios del álbum editado y actualiza su tarjeta en la vista. |
| `DELETE` | `/albums/{id}` | `DeleteAlbumHandler` | Elimina el álbum en BD (solo si sos el dueño) y devuelve `200 OK` vacío, quitando la tarjeta del DOM. |

---

## Cuentas de Usuario, Buscador y Edición de Álbumes

Se sumó un sistema simple de cuentas para que cada usuario tenga su propio perfil y colección de álbumes:

* **Registro y Login simples (`/register` y `/login`):**
  Podés crear una cuenta ingresando únicamente un **nombre de usuario** y una **contraseña**.
  > **Nota sobre la contraseña:** Los requisitos de contraseña se dejaron bien simples (mínimo 4 caracteres, sin exigir símbolos ni mayúsculas complejas) porque el objetivo es poder probar la aplicación cómodamente en un entorno de práctica. De todas formas, las contraseñas nunca se guardan en texto plano en la base de datos, sino hasheadas de forma segura (`PBKDF2-SHA256`) y la sesión se maneja con una cookie firmada.

* **Tu perfil vs. buscar otras cuentas (`/u/{username}`):**
  * Cuando entrás con tu cuenta a la página principal (`/`), ves el formulario para cargar nuevos álbumes y tu propia colección.
  * Arriba en la barra de navegación tenés un **buscador de usuarios**. Al escribir un nombre, te muestra los usuarios encontrados en tiempo real y podés entrar a ver los álbumes que cargó esa otra cuenta.
  * Cuando estás viendo el perfil de otro usuario, la página entra en **modo lectura**: se oculta el formulario de carga y tampoco aparecen los botones de editar ni eliminar, para que nadie pueda modificar la colección de otra persona.

* **Modificación y borrado solo por el dueño:**
  * En tus propios álbumes tenés el botón **Editar**, que abre una **ventana emergente flotante** sobre la página con todos los datos cargados para que puedas modificarlos y guardarlos sin recargar la pantalla.
  * Tanto la edición (`PUT /albums/{id}`) como el borrado (`DELETE /albums/{id}`) validan en la base de datos que el álbum pertenezca al usuario logueado.
