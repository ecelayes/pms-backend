# PMS Backend Core

Backend para un sistema de gestión hotelera (Property Management System) diseñado con enfoque en escalabilidad, consistencia de datos y seguridad.

El sistema utiliza una arquitectura de "Compute on Read" para el cálculo de precios y disponibilidad, evitando la desincronización de inventarios. Está construido siguiendo los principios de Clean Architecture y Multi-tenancy.

## Tecnologías

- Lenguaje: Go (Golang) 1.23+
- Base de Datos: PostgreSQL 15+
- Framework Web: Echo v4
- Driver SQL: pgx/v5
- Infraestructura: Docker & Docker Compose
- Automatización: GNU Make

## Características Principales

- Arquitectura Limpia: Separación estricta de capas (Handler, Usecase, Repository, Entity).
- Multi-Tenancy: Soporte para múltiples hoteles y dueños en la misma instancia.
- Motor de Precios Dinámico: Cálculo de tarifas en tiempo real basado en reglas y prioridades.
- Disponibilidad Transaccional: Prevención de overbooking mediante transacciones ACID y bloqueos a nivel de base de datos.
- Seguridad Avanzada: Autenticación JWT con Salt único por usuario (permite revocación inmediata de sesiones).
- Auditoría: Trazabilidad automática de creación y actualización (created_at, updated_at) y borrado lógico (Soft Delete).

## Requisitos Previos

Para ejecutar este proyecto necesitas tener instalado:

1. Go 1.23 o superior
2. Docker y Docker Compose
3. Make (generalmente incluido en Linux/Mac, o vía Chocolatey/Scoop en Windows)
4. Cliente PostgreSQL (psql) - Opcional pero recomendado para debugging

## Configuración

Crea un archivo .env en la raíz del proyecto copiando el siguiente contenido:

DB_USER=postgres
DB_PASSWORD=postgres
DB_HOST=localhost
DB_PORT=5432

DB_NAME=hotel_pms_db
DB_TEST_NAME=hotel_pms_test

PORT=8080

# No se requiere JWT_SECRET ya que usamos sales dinámicas por usuario en la DB

## Ejecución del Proyecto

El proyecto incluye un Makefile para simplificar todas las tareas comunes.

### Opción A: Ejecución con Docker (Recomendada)

Levanta la base de datos y la API en contenedores aislados. La API estará disponible en el puerto 4000.

1. Levantar servicios:
   make docker-up

2. Inicializar base de datos (Solo la primera vez o para reiniciar):
   Crea el esquema, aplica migraciones y carga datos de prueba.
   make docker-db-reset

3. Ver logs:
   make docker-logs

4. Detener servicios:
   make docker-down

### Opción B: Ejecución Local (Desarrollo)

Requiere tener una instancia de PostgreSQL corriendo localmente en el puerto 5432.

1. Preparar base de datos local:
   make db-reset

2. Iniciar el servidor (Hot reload no incluido, reiniciar manualmente):
   make run

El servidor escuchará en el puerto definido en el .env (por defecto 8081).

## Testing

El proyecto cuenta con una suite de tests de integración (End-to-End) que valida los flujos de negocio completos contra una base de datos real de prueba.

- Ejecutar todos los tests:
  make test-all

- Ejecutar solo tests unitarios (sin DB):
  make test-unit

- Ejecutar solo el test de ciclo de vida (Flujo completo):
  make test-lifecycle

## Estructura del Proyecto

/cmd
  /api          # Punto de entrada (main.go)
/internal
  /bootstrap    # Configuración e inyección de dependencias
  /entity       # Modelos de dominio y errores
  /handler      # Controladores HTTP (Entrada/Salida JSON)
  /usecase      # Lógica de negocio pura
  /repository   # Acceso a datos (SQL queries)
  /security     # Middlewares y utilidades de autenticación
/migrations     # Scripts SQL para estructura de la DB
/scripts        # Datos semilla (Seed data)
/tests          # Tests de integración E2E
