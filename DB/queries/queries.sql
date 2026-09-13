-- name: CreateAlbum :one
INSERT INTO albums (
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
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
)
RETURNING *;


-- name: GetAlbum :one
SELECT *
FROM albums
WHERE id = $1;


-- name: ListAlbums :many
SELECT *
FROM albums
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


-- name: DeleteAlbum :exec
DELETE FROM albums
WHERE id = $1;