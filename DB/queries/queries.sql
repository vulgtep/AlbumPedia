-- name: CreateUser :one
INSERT INTO users (
    username,
    password_hash
) VALUES (
    $1, $2
)
RETURNING *;

-- name: GetUserByUsername :one
SELECT *
FROM users
WHERE LOWER(username) = LOWER($1);

-- name: GetUserByID :one
SELECT *
FROM users
WHERE id = $1;

-- name: SearchUsers :many
SELECT id, username
FROM users
WHERE username ILIKE '%' || $1 || '%'
ORDER BY username
LIMIT 8;

-- name: CreateAlbum :one
INSERT INTO albums (
    user_id,
    titulo,
    genero,
    banda,
    productor,
    lista_artistas,
    portada,
    contraportada,
    duracion_minutos,
    fecha_salida,
    link_escucha,
    observacion
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
)
RETURNING *;

-- name: GetAlbum :one
SELECT *
FROM albums
WHERE id = $1;

-- name: GetAlbumByOwner :one
SELECT *
FROM albums
WHERE id = $1 AND user_id = $2;

-- name: ListAlbums :many
SELECT *
FROM albums
ORDER BY id;

-- name: ListAlbumsByUser :many
SELECT *
FROM albums
WHERE user_id = $1
ORDER BY id;

-- name: UpdateAlbum :one
UPDATE albums
SET
    titulo = $2,
    genero = $3,
    banda = $4,
    productor = $5,
    lista_artistas = $6,
    portada = $7,
    contraportada = $8,
    duracion_minutos = $9,
    fecha_salida = $10,
    link_escucha = $11,
    observacion = $12
WHERE id = $1
RETURNING *;

-- name: UpdateAlbumByOwner :one
UPDATE albums
SET
    titulo = $3,
    genero = $4,
    banda = $5,
    productor = $6,
    lista_artistas = $7,
    portada = $8,
    contraportada = $9,
    duracion_minutos = $10,
    fecha_salida = $11,
    link_escucha = $12,
    observacion = $13
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteAlbum :exec
DELETE FROM albums
WHERE id = $1;

-- name: DeleteAlbumByOwner :exec
DELETE FROM albums
WHERE id = $1 AND user_id = $2;
