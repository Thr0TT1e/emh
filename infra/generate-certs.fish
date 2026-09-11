#!/usr/bin/fish
# Установим коды цветов
set -l clr (set_color normal)  # normal
set -l err (set_color red)
set -l ok (set_color green)
set -l remote_color (set_color cyan)
set -l branch_color (set_color yellow)

# Генерация сертификатов с помощью CFSSL для Fish shell

# Создаем директорию для сертификатов
mkdir -p ../certs

# 1. Генерируем корневой CA
echo -s $ok "Генерация корневого CA..." $clr
cfssl genkey -initca cfssl/dev/ca-csr.json | cfssljson -bare ../certs/ca

# 2. Создаем отдельный CSR для серверных сертификатов
echo -s $ok "Создание CSR для серверных сертификатов..." $clr
set server_csr '{
  "CN": "EMH Certificate",
  "hosts": [
    "localhost",
    "emh-back",
    "emh-front",
    "postgres-emh",
    "minio-emh",
    "app",
    "emh",
    "127.0.0.1",
    "::1"
  ],
  "key": {
    "algo": "rsa",
    "size": 2048
  },
  "names": [
    {
      "C": "RU",
      "ST": "Moscow",
      "L": "Moscow",
      "O": "EMH",
      "OU": "IT"
    }
  ]
}'

echo $server_csr > ../certs/server-csr.json

# 3. Проверяем конфигурацию подписи
echo -s $ok "Проверка конфигурации подписи..." $clr
set ca_config '{
    "signing": {
        "default": {
            "expiry": "8760h",
            "usages": [
                "signing",
                "key encipherment",
                "server auth",
                "client auth"
            ]
        },
        "profiles": {
            "server": {
                "expiry": "8760h",
                "usages": [
                    "signing",
                    "key encipherment",
                    "server auth"
                ]
            }
        }
    }
}'

echo $ca_config > ../certs/ca-config.json

# 4. Генерируем серверные сертификаты
echo -s $ok "Генерация серверных сертификатов для хостов: " $remote_color $hosts $clr
cfssl gencert \
  -ca ../certs/ca.pem \
  -ca-key ../certs/ca-key.pem \
  -config=../certs/ca-config.json \
  -profile=server ../certs/server-csr.json \
  | cfssljson -bare ../certs/server

# 5. Устанавливаем правильные права
echo -s $ok "Установка прав доступа..." $clr
chmod 644 ../certs/*.pem
chmod 600 ../certs/*-key.pem

echo -s $ok "Сертификаты успешно сгенерированы в директории "$remote_color"../certs/" $clr
