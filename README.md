# Supply Closet

This app now supports an optional PostgreSQL database. When the `DATABASE_URL` environment
variable is set, the server connects to that database and stores inventory and issued
items in two tables: `inventory` and `issued`.

To use DigitalOcean's managed databases, create a PostgreSQL instance and set
`DATABASE_URL` to the provided connection string. On startup the server will
create tables if they do not exist and load existing data.

Run the server:

```bash
DATABASE_URL=postgres://username:password@host:port/dbname sslmode=require go run .
```

Without `DATABASE_URL`, the application continues to operate in memory only.
