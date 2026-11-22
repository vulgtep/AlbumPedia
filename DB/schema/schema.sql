CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS albums (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
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
