# Base de datos

PostgreSQL es la fuente de verdad de Atalaya. Watchman es el único servicio que
escribe directamente en ella durante la primera versión.

## Migraciones

Las migraciones SQL pertenecen a Watchman y viven en
`apps/watchman/migrations/`. Se ejecutan con `goose`:

```bash
goose -dir apps/watchman/migrations postgres "$DATABASE_URL" status
goose -dir apps/watchman/migrations postgres "$DATABASE_URL" up
```

No se deben guardar credenciales dentro de los archivos SQL ni pasar una URL real
como argumento en scripts compartidos. En CI y producción `DATABASE_URL` proviene
del secret store del entorno.
