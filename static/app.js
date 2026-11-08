// static/app.js

// Si se abre como archivo local (file://), apunta a localhost:8080; si lo sirve Go, usa ruta relativa
const API_BASE = window.location.protocol === 'file:' ? 'http://localhost:8080' : '';
const ALBUMS_URL = `${API_BASE}/albums`;

// Referencias a elementos del DOM
const albumForm = document.getElementById('album-form');
const formTitle = document.getElementById('form-title');
const albumIdInput = document.getElementById('album-id');
const btnSubmit = document.getElementById('btn-submit');
const btnCancel = document.getElementById('btn-cancel');
const btnRefresh = document.getElementById('btn-refresh');
const statusMessage = document.getElementById('status-message');
const albumsContainer = document.getElementById('albums-container');

// Inputs del formulario
const inputTitulo = document.getElementById('input-titulo');
const inputBanda = document.getElementById('input-banda');
const inputGenero = document.getElementById('input-genero');
const inputProductor = document.getElementById('input-productor');
const inputArtistas = document.getElementById('input-artistas');
const inputDuracion = document.getElementById('input-duracion');
const inputFecha = document.getElementById('input-fecha');
const inputPortada = document.getElementById('input-portada');
const inputContraportada = document.getElementById('input-contraportada');
const inputLink = document.getElementById('input-link');
const inputObservacion = document.getElementById('input-observacion');

// Helpers para serializar y deserializar tipos sql.Null* de sqlc en Go
function toNullString(value) {
    const trimmed = value.trim();
    return {
        String: trimmed,
        Valid: trimmed.length > 0
    };
}

function toNullInt32(value) {
    const parsed = parseInt(value, 10);
    const isValid = !isNaN(parsed) && parsed > 0;
    return {
        Int32: isValid ? parsed : 0,
        Valid: isValid
    };
}

function toNullTime(value) {
    const trimmed = value.trim();
    if (!trimmed) {
        return {
            Time: "0001-01-01T00:00:00Z",
            Valid: false
        };
    }
    return {
        Time: `${trimmed}T00:00:00Z`,
        Valid: true
    };
}

function fromNullString(nullStr) {
    return (nullStr && nullStr.Valid) ? nullStr.String : '';
}

function fromNullInt32(nullInt) {
    return (nullInt && nullInt.Valid) ? nullInt.Int32 : '';
}

function fromNullTime(nullTime) {
    if (nullTime && nullTime.Valid && nullTime.Time) {
        return nullTime.Time.substring(0, 10); // Formato YYYY-MM-DD
    }
    return '';
}

// Mostrar notificación temporal
function showMessage(text, type = 'success') {
    statusMessage.textContent = text;
    statusMessage.className = `status-message ${type}`;
    statusMessage.classList.remove('hidden');

    setTimeout(() => {
        statusMessage.classList.add('hidden');
    }, 4000);
}

// Escapar HTML para evitar inyección
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// Estado local de álbumes cargados
let currentAlbums = [];

// 1. Obtener y listar todos los álbumes (GET /albums)
async function fetchAlbums() {
    try {
        albumsContainer.innerHTML = '<p class="loading-text">Cargando álbumes...</p>';
        const response = await fetch(ALBUMS_URL);

        if (!response.ok) {
            throw new Error(`Error al obtener los álbumes (${response.status})`);
        }

        const albums = await response.json();
        currentAlbums = albums || [];
        renderAlbums(currentAlbums);
    } catch (error) {
        albumsContainer.innerHTML = `
            <div class="empty-state">
                <p>No se pudo conectar con la API REST (${escapeHtml(error.message)}).</p>
                <p>Verificá que el servidor Go esté corriendo en <strong>http://localhost:8080</strong>.</p>
            </div>
        `;
    }
}

// Renderizar las tarjetas de álbumes en el DOM
function renderAlbums(albums) {
    if (!albums || albums.length === 0) {
        albumsContainer.innerHTML = `
            <div class="empty-state">
                <p>Todavía no hay álbumes cargados en la base de datos.</p>
                <p>¡Usá el formulario de arriba para crear el primero!</p>
            </div>
        `;
        return;
    }

    albumsContainer.innerHTML = '';

    albums.forEach(album => {
        const card = document.createElement('article');
        card.className = 'album-card';

        const portada = fromNullString(album.portada);
        const contraportada = fromNullString(album.contraportada);
        const productor = fromNullString(album.productor);
        const artistas = fromNullString(album.lista_artistas);
        const duracion = fromNullInt32(album.duracion_minutos);
        const fecha = fromNullTime(album.fecha_salida);
        const link = fromNullString(album.link_escucha);
        const observacion = fromNullString(album.observacion);

        let coverHtml = '';
        if (portada || contraportada) {
            const initialImg = portada || contraportada;
            const initialLabel = portada ? 'Portada' : 'Contraportada';
            const hasBoth = Boolean(portada && contraportada);

            coverHtml = `
                <div class="cover-wrapper"
                     data-portada="${escapeHtml(portada)}"
                     data-contraportada="${escapeHtml(contraportada)}"
                     data-showing="${portada ? 'portada' : 'contraportada'}">
                    <img src="${escapeHtml(initialImg)}" alt="${initialLabel} de ${escapeHtml(album.titulo)}" class="album-cover" onerror="this.closest('.cover-wrapper').style.display='none'">
                    <span class="cover-label">${initialLabel}</span>
                    ${hasBoth ? `
                        <button type="button" class="cover-arrow cover-arrow-left" title="Cambiar entre Portada y Contraportada" aria-label="Imagen anterior">&#10094;</button>
                        <button type="button" class="cover-arrow cover-arrow-right" title="Cambiar entre Portada y Contraportada" aria-label="Siguiente imagen">&#10095;</button>
                    ` : ''}
                </div>
            `;
        }

        let extraDetails = '';
        if (productor) extraDetails += `<span><strong>Productor:</strong> ${escapeHtml(productor)}</span>`;
        if (artistas) extraDetails += `<span><strong>Artistas:</strong> ${escapeHtml(artistas)}</span>`;
        if (duracion) extraDetails += `<span><strong>Duración:</strong> ${escapeHtml(String(duracion))} min</span>`;
        if (fecha) extraDetails += `<span><strong>Lanzamiento:</strong> ${escapeHtml(fecha)}</span>`;

        card.innerHTML = `
            <div>
                ${coverHtml}
                <div class="album-body">
                    <div class="album-header">
                        <h3 class="album-title">${escapeHtml(album.titulo)}</h3>
                        <span class="album-badge">${escapeHtml(album.genero)}</span>
                    </div>
                    <p class="album-band">${escapeHtml(album.banda)}</p>
                    ${extraDetails ? `<div class="album-details">${extraDetails}</div>` : ''}
                    ${observacion ? `<p class="album-observation">"${escapeHtml(observacion)}"</p>` : ''}
                    ${link ? `<a href="${escapeHtml(link)}" target="_blank" rel="noopener noreferrer" class="album-link">▶ Escuchar álbum</a>` : ''}
                </div>
            </div>
        <div class="album-footer">
            <span class="album-id">#${album.id}</span>
            <div class="album-actions">
                <button class="btn btn-edit" data-id="${album.id}">Editar</button>
                <button class="btn btn-delete" data-id="${album.id}" data-titulo="${escapeHtml(album.titulo)}">Eliminar</button>
            </div>
        </div>
        `;

        albumsContainer.appendChild(card);
    });
}

// 3. Implementación del Borrado, Edición y Cambio de Portada/Contraportada usando Delegación de Eventos
albumsContainer.addEventListener('click', (event) => {
    const arrowBtn = event.target.closest('.cover-arrow');
    if (arrowBtn) {
        const wrapper = arrowBtn.closest('.cover-wrapper');
        if (!wrapper || wrapper.dataset.animating === 'true') return;

        const img = wrapper.querySelector('.album-cover');
        const label = wrapper.querySelector('.cover-label');
        const currentShowing = wrapper.dataset.showing;
        const isNext = arrowBtn.classList.contains('cover-arrow-right');
        const slideOffset = isNext ? -28 : 28;

        const nextShowing = currentShowing === 'portada' ? 'contraportada' : 'portada';
        const nextSrc = nextShowing === 'contraportada'
            ? wrapper.dataset.contraportada
            : wrapper.dataset.portada;
        const nextLabel = nextShowing === 'contraportada' ? 'Contraportada' : 'Portada';

        wrapper.dataset.animating = 'true';

        // Fase 1: Animación de salida con Web Animations API (JS nativo)
        const outAnim = img.animate(
            [
                { opacity: 1, transform: 'translateX(0) scale(1)' },
                { opacity: 0, transform: `translateX(${slideOffset}px) scale(0.95)` }
            ],
            { duration: 180, easing: 'ease-in', fill: 'forwards' }
        );

        outAnim.onfinish = () => {
            img.src = nextSrc;
            wrapper.dataset.showing = nextShowing;
            if (label) label.textContent = nextLabel;

            // Fase 2: Animación de entrada desde el lado opuesto
            const inAnim = img.animate(
                [
                    { opacity: 0, transform: `translateX(${-slideOffset}px) scale(0.95)` },
                    { opacity: 1, transform: 'translateX(0) scale(1)' }
                ],
                { duration: 220, easing: 'ease-out', fill: 'forwards' }
            );

            inAnim.onfinish = () => {
                wrapper.dataset.animating = 'false';
            };
        };
        return;
    }

    const deleteBtn = event.target.closest('.btn-delete');
    if (deleteBtn) {
        const id = deleteBtn.dataset.id;
        const titulo = deleteBtn.dataset.titulo || `#${id}`;
        deleteAlbum(id, titulo);
        return;
    }

    const editBtn = event.target.closest('.btn-edit');
    if (editBtn) {
        const id = Number(editBtn.dataset.id);
        const album = currentAlbums.find(a => a.id === id);
        if (album) {
            loadAlbumIntoForm(album);
        }
    }
});

// 2. Crear o Actualizar un álbum (POST /albums o PUT /albums/{id})
albumForm.addEventListener('submit', async (event) => {
    event.preventDefault();

    const titulo = inputTitulo.value.trim();
    const banda = inputBanda.value.trim();
    const genero = inputGenero.value.trim();

    if (!titulo || !banda || !genero) {
        showMessage('Por favor completá los campos obligatorios: Título, Banda y Género.', 'error');
        return;
    }

    const payload = {
        titulo,
        banda,
        genero,
        productor: toNullString(inputProductor.value),
        lista_artistas: toNullString(inputArtistas.value),
        portada: toNullString(inputPortada.value),
        contraportada: toNullString(inputContraportada.value),
        duracion_minutos: toNullInt32(inputDuracion.value),
        fecha_salida: toNullTime(inputFecha.value),
        link_escucha: toNullString(inputLink.value),
        observacion: toNullString(inputObservacion.value)
    };

    const editingId = albumIdInput.value;
    const isEditing = Boolean(editingId);
    const url = isEditing ? `${ALBUMS_URL}/${editingId}` : ALBUMS_URL;
    const method = isEditing ? 'PUT' : 'POST';

    try {
        btnSubmit.disabled = true;
        const response = await fetch(url, {
            method,
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(payload)
        });

        if (!response.ok) {
            const errorText = await response.text();
            throw new Error(errorText || `Error HTTP ${response.status}`);
        }

        showMessage(
            isEditing ? '¡Álbum actualizado correctamente!' : '¡Álbum creado exitosamente!',
            'success'
        );

        resetForm();
        await fetchAlbums();
    } catch (error) {
        showMessage(`No se pudo guardar el álbum: ${error.message}`, 'error');
    } finally {
        btnSubmit.disabled = false;
    }
});

// 3. Cargar datos de un álbum en el formulario para editarlo
function loadAlbumIntoForm(album) {
    albumIdInput.value = album.id;
    inputTitulo.value = album.titulo || '';
    inputBanda.value = album.banda || '';
    inputGenero.value = album.genero || '';

    inputProductor.value = fromNullString(album.productor);
    inputArtistas.value = fromNullString(album.lista_artistas);
    inputDuracion.value = fromNullInt32(album.duracion_minutos);
    inputFecha.value = fromNullTime(album.fecha_salida);
    inputPortada.value = fromNullString(album.portada);
    inputContraportada.value = fromNullString(album.contraportada);
    inputLink.value = fromNullString(album.link_escucha);
    inputObservacion.value = fromNullString(album.observacion);

    formTitle.textContent = `Editar Álbum #${album.id}`;
    btnSubmit.textContent = 'Actualizar Álbum';
    btnCancel.classList.remove('hidden');

    window.scrollTo({ top: 0, behavior: 'smooth' });
}

// Limpiar el formulario y volver a modo creación
function resetForm() {
    albumForm.reset();
    albumIdInput.value = '';
    formTitle.textContent = 'Agregar Nuevo Álbum';
    btnSubmit.textContent = 'Guardar Álbum';
    btnCancel.classList.add('hidden');
}

// 4. Eliminar un álbum (DELETE /albums/{id})
async function deleteAlbum(id, titulo) {
    const confirmacion = window.confirm(`¿Estás seguro de que querés eliminar el álbum "${titulo}"?`);
    if (!confirmacion) return;

    try {
        const response = await fetch(`${ALBUMS_URL}/${id}`, {
            method: 'DELETE'
        });

        if (!response.ok) {
            throw new Error(`Error al eliminar (${response.status})`);
        }

        showMessage(`Álbum "${titulo}" eliminado correctamente.`, 'success');
        if (albumIdInput.value === String(id)) {
            resetForm();
        }
        await fetchAlbums();
    } catch (error) {
        showMessage(`Error al eliminar el álbum: ${error.message}`, 'error');
    }
}

// Eventos adicionales
btnCancel.addEventListener('click', resetForm);
btnRefresh.addEventListener('click', fetchAlbums);

// Inicializar al cargar la página
document.addEventListener('DOMContentLoaded', fetchAlbums);

