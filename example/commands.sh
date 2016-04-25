#!/usr/bin/env bash

# для serve на Go машинках нужен будет /etc/serve/config.yml,
# в котором будут адреса марафона, реестра пакетов и прочие конфиги

# - собираем пакет и загружаем артефакты в репозиторий (apt, task-regestry, maven, etc)
serve build --build-number '34' --branch 'master'

# - запускаем новую версию,
# - дожидаемся появления в консуле
serve deploy --env 'qa' --build-number '34' --branch 'master'

# - находим сервис в консуле
# - добавляем ему роутинг-параметры чтобы он в nginx попал
# - удаляем предыдущую версию с таким же staging из консула
# - стопаем предыдущую версию с таким же staging в марафоне (через 3 минуты)
serve release --env 'qa' --build-number '34' --branch 'master'   # --branch опционально поле, по-умолчанию master

serve release --env 'live' --staging 'stage' --build-number '34'
serve release --env 'live' --staging 'live' --build-number '34'


# скрипт-wrapper для регистрации
serve consul supervisor \
  --name 'forgame-api3' \
  --version '1.0.34' \
  --host 'api.4gametest.com' \
  --location '/v3/' \
  --staging 'live' \
  --port 12073
  start bin/start -Xmx521m ...


# OPTIONAL: вариант без использования manifest.yml

serve marathon deploy-site \
  --marathon 'mesos1-q.qa.inn.ru' \
  --name 'forgame-api3' \
  --version '1.0.34' \
  --host 'api.4gametest.com' \
  --location '/v3/' \
  --staging 'live' \
  --instances 2 \
  --mem 512 \
  --constraints 'cpb' \
  --envs '{"ENV":"qa"}'

serve marathon release-site \
  --marathon 'mesos1-q.qa.inn.ru' \
  --name 'forgame-api3' \
  --version '1.0.34'

serve marathon deploy-task \
  --marathon 'mesos1-q.qa.inn.ru' \
  --name 'forgame-api3' \
  --version '1.0.34' \
  --instances 2 \
  --mem 512 \
  --constraints 'cpb' \
  --envs '{"ENV":"qa"}'
