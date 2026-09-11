#!/bin/bash
set -e

clrReset="\033[0m"
clrGreen="\033[0;32m"

if [ -z "$POSTGRES_PASSWORD" ] || [ -z "$POSTGRES_USER" ] || [ -z "$POSTGRES_DB" ]; then
    echo "ERROR: POSTGRES_PASSWORD, POSTGRES_USER, and POSTGRES_DB must be set"
    exit 1
fi

# Установка локали
echo -e "$clrGreenУстановка локали$clrReset"
apt-get update && apt-get install -y locales && rm -rf /var/lib/apt/lists/* && localedef -i ru_RU -c -f UTF-8 -A /usr/share/locale/locale.alias ru_RU.UTF-8

# Настройка прав на сертификаты
echo -e "$clrGreenНастройка прав на сертификаты$clrReset"
cp -r /app/certs /var/lib/postgresql/certs/
chown -R postgres:postgres /var/lib/postgresql/certs/
chmod 600 /var/lib/postgresql/certs/server-key.pem
chmod 644 /var/lib/postgresql/certs/server.pem
chmod 644 /var/lib/postgresql/certs/ca.pem

# Функция для инициализации БД
init_postgres() {
    echo -e "$clrGreenИнициализация БД$clrReset"
    su postgres -c "initdb -D $PGDATA \
        --auth-host=md5 \
        --auth-local=md5 \
        --locale=ru_RU.utf8 \
        --encoding=UTF8 \
        --lc-collate=ru_RU.utf8 \
        --lc-ctype=ru_RU.utf8 \
        --username=postgres \
        --pwfile=<(echo "$POSTGRES_PASSWORD")"

    echo -e "$clrGreenНастройка конфигурации$clrReset"
    cat >> $PGDATA/postgresql.conf <<EOF
ssl = on
ssl_cert_file = '/var/lib/postgresql/certs/server.pem'
ssl_key_file = '/var/lib/postgresql/certs/server-key.pem'
ssl_ca_file = '/var/lib/postgresql/certs/ca.pem'
listen_addresses = '*'
port = 5432
EOF

    echo -e "$clrGreenНастройка pg_hba.conf$clrReset"
    echo "hostssl all all all md5" >> $PGDATA/pg_hba.conf

    echo -e "$clrGreenЗапуск сервера для настройки$clrReset"
    su postgres -c "pg_ctl -D $PGDATA -w start"

    echo -e "$clrGreenСоздание пользователя и БД$clrReset"
    PGPASSWORD="$POSTGRES_PASSWORD" psql -h localhost -U postgres -c "CREATE USER $POSTGRES_USER WITH PASSWORD '$POSTGRES_PASSWORD';"
    PGPASSWORD="$POSTGRES_PASSWORD" psql -h localhost -U postgres -c "CREATE DATABASE $POSTGRES_DB OWNER $POSTGRES_USER;"

    echo -e "$clrGreenОстановка сервера$clrReset"
    su postgres -c "pg_ctl -D $PGDATA -w stop"
}

# Инициализация базы данных если нужно
if [ ! -s "$PGDATA/PG_VERSION" ]; then
    init_postgres
fi

echo -e "$clrGreenСоздание флага готовности$clrReset"
touch /var/lib/postgresql/initialized

echo -e "$clrGreenЗапуск PostgreSQL в foreground$clrReset"
exec su postgres -c "postgres -D $PGDATA"
