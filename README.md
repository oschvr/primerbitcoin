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


## Prom + Grafana

We'll install prometheus + grafana on a raspberry pi 4

```
# run as root/sudo user
sudo su

# Get latest binaries for arm
VERSION="2.47.2"
wget "https://github.com/prometheus/prometheus/releases/download/v${VERSION}/prometheus-${VERSION}.linux-arm64.tar.gz"
tar xfz "prometheus-${VERSION}.linux-arm64.tar.gz"
rm "prometheus-${VERSION}.linux-arm64.tar.gz"
mv "prometheus-${VERSION}.linux-arm64/" "/usr/local/lib/prometheus/"

# create systemd unit
cat << EOF > /etc/systemd/system/prometheus.service   
[Unit]
Description=Prometheus Server
Documentation=https://prometheus.io/docs/introduction/overview/
After=network-online.target

[Service]
User=saturn
Restart=on-failure

#Change this line if Prometheus is somewhere different
ExecStart=/usr/local/lib/prometheus/prometheus \
  --config.file=/usr/local/lib/prometheus/prometheus.yml \
  --storage.tsdb.path=/usr/local/lib/prometheus/data

[Install]
WantedBy=multi-user.target 
EOF

# change ownership of user
chown -R saturn: /usr/local/lib/prometheus

# reload
systemctl daemon-reload

# enable and start
systemctl enable prometheus
systemctl start prometheus

# add primerbitcoin metrics
# /usr/local/lib/prometheus/prometheus.yml
----
scrape_configs:
  ...
  - job_name: "primerbitcoin"
    metrics_path: '/metrics'
    static_configs:
      - targets: ["metrics.primerbitcoin.com"]
    basic_auth:
      username: 'primerbitcoin'
      password: 'cccccbhlitbuchfgjhenfuvberlchkgekctcuefjlgbi'
----
```


`blackbox_exporter-0.24.0.linux-arm64.tar.gz`
`node_exporter-1.6.1.linux-arm64.tar.gz`
``