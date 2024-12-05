## **Descripción del Proyecto**
Este proyecto es una implementación de una arquitectura de microservicios en **Go (Golang)** que incluye tres servicios principales:

1. **User-Service**: Gestión de usuarios, incluyendo su almacenamiento en SQLite y cacheo en Redis.
2. **Tweets-Service**: Manejo de tweets, con persistencia en SQLite y almacenamiento en Redis para lecturas rápidas.
3. **Timeline-Service**: Generación y manejo de timelines de usuarios utilizando Redis para almacenamiento y procesamiento eficiente.

Todos los servicios están diseñados siguiendo principios de **Clean Architecture** para garantizar modularidad y escalabilidad.

---

## **Tecnologías utilizadas**
- **Go (Golang)**: Lenguaje principal de implementación.
- **Redis**: Almacenamiento en memoria para caché y procesamiento de datos temporales.
- **SQLite**: Base de datos ligera para persistencia.
- **Docker & Docker Compose**: Orquestación de contenedores para desplegar los servicios.

---

## **Estructura de los Servicios**
Cada servicio tiene la misma estructura base:
- **`cmd`**: Contiene el archivo principal del servicio.
- **`config`**: Configuración del servicio (archivos YAML y estructuras en Go).
- **`internal`**:
  - **application**: Lógica de negocio.
  - **domain**: Modelos y entidades del dominio.
  - **infrastructure**: Manejo de base de datos, controladores HTTP, entre otros.
- **`seeder`**: Generación de datos fake para pruebas (en User-Service y Tweets-Service).

---

# User-Service: Rutas disponibles

POST /users
- Función: Crear un nuevo usuario en el sistema.
- Autenticación: No requerida.

POST /users/:id/follow
- Función: Permitir que un usuario autenticado siga a otro usuario identificado por `id`.
- Autenticación: Requerida mediante un header con el formato:
  User-ID: 2a42c7ae-7f78-4e36-8358-902342fe23f1

POST /users/:id/unfollow
- Función: Permitir que un usuario autenticado deje de seguir a otro usuario identificado por `id`.
- Autenticación: Requerida mediante un header con el formato:
  User-ID: 2a42c7ae-7f78-4e36-8358-902342fe23f1

# Tweets-Service: Rutas disponibles

POST /tweets
- Función: Crear un tweet para un usuario autenticado.
- Autenticación: Requerida mediante un header con el formato:
  User-ID: 2a42c7ae-7f78-4e36-8358-902342fe23f1
- Notas: Asegurar que el tweet no supere los 280 caracteres.

DELETE /tweets/:id
- Función: Permitir que un usuario autenticado elimine uno de sus tweets.
- Autenticación: Requerida mediante un header con el formato:
  User-ID: 2a42c7ae-7f78-4e36-8358-902342fe23f1

# Timeline-Service: Rutas disponibles

GET /paginate
- Función: Obtener un timeline paginado con los tweets de los usuarios seguidos por un usuario autenticado.
- Autenticación: Requerida mediante un header con el formato:
  User-ID: 2a42c7ae-7f78-4e36-8358-902342fe23f1
- Notas: Optimizar paginación utilizando Redis.


## **Cómo levantar el proyecto**
1. **Requisitos previos**:
   - Tener instalado **Docker** y **Docker Compose**.

2. **Iniciar los servicios**:
   Ejecuta el siguiente comando desde el directorio raíz del proyecto:
   ```bash
   docker-compose up --build