#!/bin/sh

echo "Current env is $APP_ENV"

if [[ "$APP_ENV" == "development" ]]; then
  npm install
fi

# executer des migrations ?

exec "$@"