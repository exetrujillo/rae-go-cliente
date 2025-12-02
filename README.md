# RAE API Client (Go)

Cliente optimizado de la API de la Real Academia Española (RAE) escrito en Go. Este servidor HTTP actúa como proxy JSON para el diccionario de la RAE.

## Características

- ✨ **Respuestas JSON limpias** (sin HTML)
- 🚀 **Eficiente y ligero** (escrito en Go)
- 🐳 **Dockerizado** para fácil despliegue
- 🔒 **Bypass de Cloudflare** usando TLS fingerprinting
- 🧹 **Procesamiento automático** de respuestas JSONP

## Instalación y Uso

### Con Docker (Recomendado)

```bash
# Construir la imagen
docker build -t rae-client .

# Ejecutar el contenedor
docker run -d -p 8080:8080 --restart always --name rae-server rae-client
```

### Sin Docker

Requiere Go 1.21+

```bash
go mod tidy
go run cmd/server/main.go
```

El servidor estará disponible en `http://localhost:8080`

## Endpoints de la API

### 📅 Palabra del Día

Obtiene la palabra del día del diccionario de la RAE.

**Endpoint:** `GET /wotd`

**Ejemplo:**
```bash
curl http://localhost:8080/wotd
```

**Respuesta:**
```json
{
  "header": "venablo",
  "id": "bVfXHvn"
}
```

**Campos:**
- `header`: La palabra del día
- `id`: ID de la palabra para obtener su definición completa

---

### 🔍 Buscar Palabra

Busca una palabra en el diccionario y devuelve resultados coincidentes.

**Endpoint:** `GET /search?w={palabra}`

**Parámetros:**
- `w` (requerido): La palabra a buscar

**Ejemplo:**
```bash
curl "http://localhost:8080/search?w=hola"
```

**Respuesta:**
```json
{
  "approx": 0,
  "res": [
    {
      "header": "hola",
      "id": "KYtLWBc",
      "grp": 0
    }
  ]
}
```

**Campos:**
- `approx`: Indica si la búsqueda es aproximada (0 = exacta, 1 = aproximada)
- `res`: Array de resultados
  - `header`: Palabra encontrada
  - `id`: ID de la palabra (usar con `/fetch` para obtener definición)
  - `grp`: Grupo de la palabra

---

### 📖 Obtener Definición

Obtiene la definición completa de una palabra usando su ID (obtenido de `/search`).

**Endpoint:** `GET /fetch?id={id}`

**Parámetros:**
- `id` (requerido): ID de la palabra
- `conjugaciones` (opcional): `true` para incluir tablas de conjugación (solo verbos)

**Ejemplo:**
```bash
curl "http://localhost:8080/fetch?id=KYtLWBc"
```

**Respuesta:**
```json
{
  "id": "5rTR6tT",
  "encabezado": "bonito",
  "etimologia": "Del b. lat. boniton.",
  "definiciones": [
    {
      "tipo": "nombre masculino",
      "definicion": "Pez teleósteo comestible, parecido al atún, pero más pequeño.",
      "sinonimos": [
        "atún",
        "biza",
        "bonítalo"
      ]
    }
  ]
}
```

> **Nota:** Los campos `sinonimos` y `antonimos` solo aparecen si la palabra tiene sinónimos o antónimos registrados en la RAE.

**Campos:**
- `id`: ID de la palabra
- `encabezado`: Palabra principal
- `etimologia`: Origen de la palabra (si existe)
- `definiciones`: Lista de definiciones
  - `tipo`: Tipo gramatical (ej. "nombre masculino")
  - `definicion`: Texto de la definición
  - `sinonimos`: Lista de sinónimos (si existen)
  - `antonimos`: Lista de antónimos (si existen)
  - `conjugaciones`: Lista de conjugaciones (si se solicita con `conjugaciones=true`)
    - `modo`: Modo verbal (Indicativo, Subjuntivo, etc.)
    - `tiempos`: Mapa de tiempos verbales y sus formas

---

## Ejemplos de Integración

Aquí tienes ejemplos de cómo consumir la API para realizar un flujo completo: **Buscar una palabra y obtener su definición**.

### Python (usando `requests`)

```python
import requests

BASE_URL = "http://localhost:8080"

def obtener_definicion(palabra):
    # 1. Buscar la palabra para obtener su ID
    print(f"Buscando '{palabra}'...")
    resp_busqueda = requests.get(f"{BASE_URL}/search", params={"w": palabra})
    datos_busqueda = resp_busqueda.json()

    if not datos_busqueda.get("res"):
        print("Palabra no encontrada.")
        return

    # Tomamos el primer resultado
    primer_resultado = datos_busqueda["res"][0]
    id_palabra = primer_resultado["id"]
    print(f"ID encontrado: {id_palabra} ({primer_resultado['header']})")

    # 2. Obtener la definición completa usando el ID
    resp_definicion = requests.get(f"{BASE_URL}/fetch", params={"id": id_palabra})
    datos_definicion = resp_definicion.json()

    # Imprimir resultados
    print(f"\nDefiniciones de '{datos_definicion['encabezado']}':")
    for d in datos_definicion.get("definiciones", []):
        print(f"- [{d['tipo']}] {d['definicion']}")
        
    if datos_definicion.get("sinonimos"):
        print(f"Sinónimos: {', '.join(datos_definicion['sinonimos'])}")

# Ejecutar
obtener_definicion("bonito")
```

### JavaScript (Node.js / Navegador)

```javascript
const BASE_URL = "http://localhost:8080";

async function obtenerDefinicion(palabra) {
    try {
        // 1. Buscar la palabra
        console.log(`Buscando '${palabra}'...`);
        const respBusqueda = await fetch(`${BASE_URL}/search?w=${encodeURIComponent(palabra)}`);
        const datosBusqueda = await respBusqueda.json();

        if (!datosBusqueda.res || datosBusqueda.res.length === 0) {
            console.log("Palabra no encontrada.");
            return;
        }

        // Tomamos el primer resultado
        const primerResultado = datosBusqueda.res[0];
        const idPalabra = primerResultado.id;
        console.log(`ID encontrado: ${idPalabra} (${primerResultado.header})`);

        // 2. Obtener la definición
        const respDefinicion = await fetch(`${BASE_URL}/fetch?id=${encodeURIComponent(idPalabra)}`);
        const datosDefinicion = await respDefinicion.json();

        // Imprimir resultados
        console.log(`\nDefiniciones de '${datosDefinicion.encabezado}':`);
        datosDefinicion.definiciones.forEach(d => {
            console.log(`- [${d.tipo}] ${d.definicion}`);
        });

        if (datosDefinicion.sinonimos) {
            console.log(`Sinónimos: ${datosDefinicion.sinonimos.join(", ")}`);
        }

    } catch (error) {
        console.error("Error:", error);
    }
}

// Ejecutar
obtenerDefinicion("bonito");
```

### R (usando `httr` y `jsonlite`)

```r
library(httr)
library(jsonlite)

base_url <- "http://localhost:8080"

obtener_definicion <- function(palabra) {
  # 1. Buscar la palabra
  message(paste("Buscando", palabra, "..."))
  resp_busqueda <- GET(paste0(base_url, "/search"), query = list(w = palabra))
  datos_busqueda <- fromJSON(content(resp_busqueda, "text", encoding = "UTF-8"))
  
  if (length(datos_busqueda$res) == 0) {
    message("Palabra no encontrada.")
    return(NULL)
  }
  
  # Tomamos el primer resultado
  # Nota: En R los índices empiezan en 1
  primer_resultado <- datos_busqueda$res[1,]
  id_palabra <- primer_resultado$id
  message(paste("ID encontrado:", id_palabra, "(", primer_resultado$header, ")"))
  
  # 2. Obtener la definición
  resp_definicion <- GET(paste0(base_url, "/fetch"), query = list(id = id_palabra))
  datos_definicion <- fromJSON(content(resp_definicion, "text", encoding = "UTF-8"))
  
  # Imprimir resultados
  cat(paste0("\nDefiniciones de '", datos_definicion$encabezado, "':\n"))
  definitions <- datos_definicion$definiciones
  
  # Iterar sobre definiciones (si es un data.frame)
  if (is.data.frame(definitions)) {
    for(i in 1:nrow(definitions)) {
      cat(paste0("- [", definitions$tipo[i], "] ", definitions$definicion[i], "\n"))
    }
  }
  
  if (!is.null(datos_definicion$sinonimos)) {
    cat(paste("Sinónimos:", paste(datos_definicion$sinonimos, collapse = ", "), "\n"))
  }
}

# Ejecutar
obtener_definicion("bonito")
```

### 🎲 Palabra Aleatoria

Obtiene una palabra aleatoria del diccionario.

**Endpoint:** `GET /random`

**Ejemplo:**
```bash
curl http://localhost:8080/random
```

**Nota:** Este endpoint devuelve HTML en lugar de JSON. Si necesitas JSON, puedes usar el ID de la palabra devuelta con el endpoint `/fetch`.

---

### 💡 Autocompletado / Sugerencias

Obtiene sugerencias de palabras basadas en una entrada parcial.

**Endpoint:** `GET /keys?q={consulta}`

**Parámetros:**
- `q` (requerido): Texto parcial para autocompletar

**Ejemplo:**
```bash
curl "http://localhost:8080/keys?q=hol"
```

**Respuesta:**
```json
[
  "hola",
  "holán",
  "holandes",
  "holandesa"
]
```

---

### 🔤 Buscar Anagramas

Busca anagramas de una palabra.

**Endpoint:** `GET /anagram?w={palabra}`

**Parámetros:**
- `w` (requerido): La palabra para buscar anagramas

**Ejemplo:**
```bash
curl "http://localhost:8080/anagram?w=amor"
```

**Respuesta:**
```json
[
  "amor",
  "armo",
  "mora",
  "omar",
  "roma"
]
```

---


## Detalles Técnicos

### Stack Tecnológico
- **Lenguaje:** Go 1.21
- **Servidor HTTP:** net/http estándar
- **Cliente HTTP:** tls-client (para bypass de Cloudflare)
- **Contenedor:** Docker multi-stage build

### Procesamiento de Respuestas
El cliente procesa automáticamente las respuestas de la RAE:
1. Desenvuelve callbacks JSONP (`json(...)` y `jsonp123(...)`)
2. Limpia entidades HTML (`&#xE1;` → `á`)
3. Elimina etiquetas superíndice
4. Devuelve JSON limpio

### API Upstream
- **URL Base:** `https://dle.rae.es/data/`
- **Autenticación:** Token de la aplicación móvil de la RAE
- **TLS:** Perfil Safari iOS 16.0 para evitar detección de Cloudflare

## Configuración

### Puerto
Por defecto el servidor escucha en el puerto `8080`. Puedes cambiarlo usando la variable de entorno `PORT`:

```bash
docker run -d -p 3000:3000 -e PORT=3000 --name rae-server rae-client
```

## Limitaciones

- El endpoint `/random` devuelve HTML en lugar de JSON (limitación de la API upstream)
- Las respuestas dependen de la disponibilidad de la API de la RAE
- El bypass de Cloudflare puede dejar de funcionar si la RAE actualiza su protección

## Estructura del Proyecto

```
rae-go/
├── cmd/
│   └── server/
│       └── main.go          # Punto de entrada del servidor HTTP
├── pkg/
│   └── rae/
│       └── client.go        # Cliente de la API con bypass de Cloudflare
├── Dockerfile               # Build multi-stage optimizado
├── go.mod                   # Dependencias del proyecto
└── README.md               # Este archivo
```

## Contribuir

Las contribuciones son bienvenidas. Por favor:
1. Haz fork del proyecto
2. Crea una rama para tu feature
3. Haz commit de tus cambios
4. Envía un pull request

## Licencia

MIT

## Créditos

Este proyecto se inspira en:
- [RAE-API](https://github.com/account0123/RAE-API) por account0123
- [mgp25/RAE-API](https://github.com/mgp25/RAE-API) por mgp25
