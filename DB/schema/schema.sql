CREATE TABLE albums (
    id SERIAL PRIMARY KEY,
    titulo VARCHAR(150) NOT NULL,
    genero VARCHAR(100) NOT NULL,
    banda VARCHAR(150) NOT NULL,
    productor VARCHAR(150),
    lista_artistas VARCHAR(150),
    portada VARCHAR(255),
    contraportada VARCHAR(255),
    duracion_minutos INTEGER,
    fecha_salida DATE,
    link_escucha VARCHAR(500),
    observacion VARCHAR(1000)
);