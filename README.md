# box - make dumps of databases

Make database dumps and store daily, weekly and monthly. Supported databases:

* PostgreSQL 9-15
* MongoDB 2.6-4.0
* MongoDB 4.0-6.0
* Firebird 2.5

## Build

```bash
go build -a -o app . 
```

## Configuration

Configuration stored in `application.yml`. See `application.sample.yml` for reference.
