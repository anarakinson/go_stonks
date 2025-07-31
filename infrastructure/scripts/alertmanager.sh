#!/bin/sh
set -e

# Генерация конфига
sed -e "s|\\\${SMTP_USER}|${SMTP_USER}|g" \
    -e "s|\\\${SMTP_PASSWORD}|${SMTP_PASSWORD}|g" \
    -e "s|\\\${SMTP_RECEIVER}|${SMTP_RECEIVER}|g" \
    /etc/alertmanager/template.yml > /etc/alertmanager/alertmanager.yml

# Запуск основной команды
exec "$@"