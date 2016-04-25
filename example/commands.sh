#!/usr/bin/env bash

# для serve на Go машинках нужен будет /etc/serve/config.yml,
# в котором будут адреса марафона, реестра пакетов и прочие конфиги

# - собираем пакет и загружаем артефакты в репозиторий (apt, task-regestry, maven, etc)
serve build --branch 'master' --build-number '34'

# - запускаем новую версию, дожидаемся появления в консуле (--branch 'master' — опционально поле, 'master' по-умолчанию)
# - находим предыдущую версию и убираем ее сначала из консула, потом из марафона (через 3 минуты)
serve deploy --env 'qa' --branch 'master' --build-number '34'

# - меняем staging:stage --> staging:live в консуле (todo: надо подумать как находить полную версию только по build-number)
# - удаляем предыдущий live из консула
# - стопаем предыдущий live в фарафоне (через 3 минуты)
serve release --env 'live' --build-number '34'


# скрипт-wrapper для регистрации
serve consul supervisor \
  --name 'forgame-api3' \
  --version '1.0.34' \
  --host 'api.4gametest.com' \
  --location '/v3/' \
  --staging 'live' \
  --port 12073
  start bin/start -Xmx521m ...


# вариант без использования manifest.yml

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
