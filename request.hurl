# 1. Crear un nuevo álbum
POST http://localhost:8080/albums
Content-Type: application/json
{
  "titulo": "Artaud",
  "genero": "Rock",
  "banda": "Pescado Rabioso"
}
HTTP 201
[Asserts]
jsonpath "$.titulo" == "Artaud"
jsonpath "$.id" isInteger

# 2. Listar todos los álbumes
GET http://localhost:8080/albums
HTTP 200
[Asserts]
jsonpath "$" count > 0

# 3. Obtener un álbum específico asumimos que el 1 es el que acabamos de crear
GET http://localhost:8080/albums/1
HTTP 200
[Asserts]
jsonpath "$.titulo" == "Artaud"
jsonpath "$.banda" == "Pescado Rabioso"

# 4. Actualizar el álbum
PUT http://localhost:8080/albums/1
Content-Type: application/json
{
  "titulo": "Artaud (Remasterizado)",
  "genero": "Rock Clásico",
  "banda": "Pescado Rabioso"
}
HTTP 200
[Asserts]
jsonpath "$.titulo" == "Artaud (Remasterizado)"
jsonpath "$.genero" == "Rock Clásico"

# 5. Eliminar el álbum
DELETE http://localhost:8080/albums/1
HTTP 204