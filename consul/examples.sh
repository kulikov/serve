#!/bin/bash

serve consul supervisor \
  --name='forgame-api3' \
  --version='1.0.34' \
  --host='api.4gametest.com' \
  --location='/v3/' \
  --staging='live' \
  bin/start

serve marathon deploy-site \
  --marathon='mesos1-q.qa.inn.ru' \
  --name='forgame-api3' \
  --version='1.0.34' \
  --host='api.4gametest.com' \
  --location='/v3/' \
  --staging='live' \
  --instances=2 \
  --mem=512 \
  --constraints='cpb' \
  --envs='{"ENV":"qa"}'

serve marathon release-site \
  --marathon='mesos1-q.qa.inn.ru' \
  --name='forgame-api3' \
  --version='1.0.34'

serve marathon deploy-task \
  --marathon='mesos1-q.qa.inn.ru' \
  --name='forgame-api3' \
  --version='1.0.34' \
  --instances=2 \
  --mem=512 \
  --constraints='cpb' \
  --envs='{"ENV":"qa"}'