#!/bin/bash

# Скрипт инициализации Cadence сервера

set -e

echo "Starting Cadence initialization..."

# Ждем доступности PostgreSQL
echo "Waiting for PostgreSQL to be ready..."
until pg_isready -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER; do
    echo "PostgreSQL is unavailable - sleeping"
    sleep 2
done

echo "PostgreSQL is ready!"

# Инициализируем схему базы данных
echo "Initializing Cadence database schema..."
cadence-sql-tool \
    --host $POSTGRES_HOST \
    --port $POSTGRES_PORT \
    --user $POSTGRES_USER \
    --password $POSTGRES_PASSWORD \
    --database $POSTGRES_DATABASE \
    --plugin postgres \
    --schema_dir /etc/cadence/schema/postgresql/v96/cadence/versioned \
    create-schema

echo "Database schema initialized!"

# Создаем домен по умолчанию если он не существует
echo "Creating default domain..."
cadence-sql-tool \
    --host $POSTGRES_HOST \
    --port $POSTGRES_PORT \
    --user $POSTGRES_USER \
    --password $POSTGRES_PASSWORD \
    --database $POSTGRES_DATABASE \
    --plugin postgres \
    --schema_dir /etc/cadence/schema/postgresql/v96/cadence/versioned \
    --domain default \
    create-domain

echo "Default domain created!"

# Запускаем Cadence сервер
echo "Starting Cadence server..."
exec cadence-server \
    --config /etc/cadence/config/docker.yaml \
    --services frontend,matching,history,worker
