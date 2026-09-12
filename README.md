# mywebserver

Un servidor web HTTP ligero y eficiente escrito en **Go**, diseñado para servir archivos estáticos desde cualquier directorio local con registro (*logging*) de peticiones HTTP en tiempo real.

---

## 🚀 Características

- **Servidor de archivos estáticos**: Sirve sitios web, assets, HTML, CSS, JavaScript e imágenes de forma rápida.
- **Apertura automática del navegador**: Abre automáticamente el navegador predeterminado al iniciar el servidor.
- **Recarga en vivo (Live Reload)**: Recarga automáticamente el navegador en tiempo real cuando se modifican, crean o eliminan archivos en la carpeta servida.
- **Configurable por línea de comandos**: Permite especificar el puerto de escucha, el directorio raíz, y activar/desactivar la apertura del navegador y el live reload mediante flags.
- **Logging de peticiones HTTP**: Registra el método HTTP, URI, dirección IP del cliente, estado, referer y User-Agent en la consola.
- **Compilación multiplataforma**: Compatible con Linux, macOS y Windows con compilación a binarios ejecutables independientes.

---

## 📋 Requisitos Previos

- [Go](https://go.dev/) **1.26** o superior.

---

## 🛠️ Instalación y Compilación

1. **Clona el repositorio:**
   ```bash
   git clone https://github.com/juancarbajal/mywebserver.git
   cd mywebserver
   ```

2. **Compila el binario:**
   ```bash
   go build -o mywebserver .
   ```

---

## 💻 Uso

Puedes ejecutar el servidor directamente compilando el proyecto o usando `go run`.

### 1. Ejecutar con valores por defecto
Por defecto, el servidor se ejecuta en el puerto **8000** y sirve los archivos del directorio actual (`.`):

```bash
./mywebserver
```

o sin compilar previo:

```bash
go run .
```

### 2. Especificar un puerto y directorio personalizado

```bash
./mywebserver -port 8080 -dir ./test
```

### 🎛️ Opciones de Línea de Comandos (Flags)

| Flag | Descripción | Valor por defecto | Ejemplo |
| :--- | :--- | :--- | :--- |
| `-port` | Puerto TCP en el que escuchará el servidor | `8000` | `-port 3000` |
| `-dir` | Directorio local desde el cual se servirán los archivos | `.` | `-dir ./test` |
| `-open` | Abre automáticamente el navegador predeterminado | `true` | `-open=false` |
| `-reload` | Habilita la recarga automática en el navegador ante cambios | `true` | `-reload=false` |

---

## 📄 Ejemplo de Salida

Al iniciar el servidor:

```text
	Starting file server...
	Serving directory: /home/usuario/Projects/mywebserver/test
	Server running on port 8080
	Access from local machine: http://localhost:8080
	Live reload: enabled
	Use Ctrl+C to stop the server
```

Cuando se recibe una petición HTTP en la consola:

```text
GET /index.html 200 412 127.0.0.1 - Mozilla/5.0 (X11; Linux x86_64) ...
```

Cuando se modifica un archivo en la carpeta servida:

```text
[reload] File changed: index.html -> reloading browser...
```

---

## 📁 Estructura del Proyecto

```text
mywebserver/
├── main.go            # Punto de entrada, parseo de flags y configuración del servidor HTTP
├── browser.go         # Detección y apertura multiplataforma del navegador predeterminado
├── livereload.go      # Observador de archivos (fsnotify), endpoint SSE y middleware de inyección
├── log.go             # Estructura HTTPReqInfo y formateador de logs de peticiones
├── browser_test.go    # Pruebas unitarias para apertura del navegador
├── livereload_test.go # Pruebas de inyección HTML, filtrado de archivos y recarga SSE
├── test/              # Directorio de prueba con un archivo index.html de muestra
│   └── index.html
├── Makefile           # Tareas de compilación multiplataforma y ejecución
├── go.mod             # Definición del módulo de Go
├── go.sum             # Checksums de dependencias
├── LICENSE            # Licencia del proyecto (MIT)
└── README.md          # Documentación del proyecto
```

---

## 📝 Licencia

Este proyecto está bajo la Licencia **MIT**. Consulta el archivo [`LICENSE`](LICENSE) para más detalles.
