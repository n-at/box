# box - make dumps of PostgreSQL and MongoDB databases

Make database dumps and store daily, weekly and monthly dumps. Supported databases:

* PostgreSQL 9-14
* MongoDB 2.6-4.0
* MongoDB 4.0-5.0
* Firebird 2.5

## Build

```bash
go build -a -o app . 
```

## Configuration

Configuration stored in `application.yml`. See `application.sample.yml` for reference.
