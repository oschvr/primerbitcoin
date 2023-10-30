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


## NGINX
Adding nginx with basic auth to serve metrics
```
# run as root/sudo user
sudo su

# Install nginx
apt install nginx apache2-utils -y

# Adjust the firewall 
ufw app list

# Enable ssh to avoid being locked out
ufw allow 'OpenSSH'

# Allow NGINX https
ufw allow 'Nginx HTTPS'

# Enable & start nginx
systemctl status nginx

# Generate random hex password
PASS=$(openssl rand -hex 18)
echo "[primerbitcoin]: nginx basic auth password is $PASS"

# Create basic auth password 
htpasswd -b -c /etc/apache2/.htpasswd primerbitcoin $(echo $PASS)
cat /etc/apache2/.htpasswd

# Create server block
cat << 'EOF' > /etc/nginx/sites-available/primerbitcoin.com
server {
  listen          443;
  listen          [::]:443;
  server_name     primerbitcoin.com;
  location /api {
        proxy_pass http://127.0.0.1:8080/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header X-Forwarded-Host $host;
        
        auth_basic "Admin";
        auth_basic_user_file /etc/apache2/.htpasswd; 
  }
  location /metrics {
        proxy_pass http://127.0.0.1:9090/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header X-Forwarded-Host $host;
        
        auth_basic "Admin";
        auth_basic_user_file /etc/apache2/.htpasswd; 
  }
}
EOF
# Check server block
cat /etc/nginx/sites-available/primerbitcoin.com

# Enable server block with symlink
ln -s /etc/nginx/sites-available/primerbitcoin.com /etc/nginx/sites-enabled/

# Check and Reload nginx
nginx -t
systemctl restart nginx

----

## Use Let's Encrypt to secure nginx at port 443
# https://www.digitalocean.com/community/tutorials/how-to-secure-nginx-with-let-s-encrypt-on-ubuntu-22-04

# Install nginx using snap
snap install --classic certbot

# symlink certbot
ln -s /snap/bin/certbot /usr/bin/certbot





```

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
      password: 'cccccbhlitbuucecrntvuldgjerkvdtkfebjhhhfuvfd'
----
```


`blackbox_exporter-0.24.0.linux-arm64.tar.gz`
`node_exporter-1.6.1.linux-arm64.tar.gz`
``