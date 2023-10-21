# primerbitcoin

DCA bot written in Golang to buy BTC on a configurable schedule through a cronjob

### migrations

https://earthly.dev/blog/golang-sqlite/

Use package `golang-migrate` for sqlite

```shell
go get -tags 'sqlite3' -u github.com/golang-migrate/migrate/v4/cmd/migrate
```

Then create a new migration

```shell
migrate create -ext sql -dir db/migrations -seq <migration_name>
```

### User in linux

```
sudo groupadd --system primerbitcoin
sudo useradd -s /sbin/bash --system -g 
```

## Concepts
- amount refers to the quantity of minor (fiat) that is intended to be spent and traded for the quantity
- quantity refers to the quantity of major (crypto) to be bought (that is, the floating point, decimal of the amount)