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

