FROM node:22-alpine3.23

WORKDIR /app

COPY . /app

RUN npm install

RUN npm run build

COPY .docker/next/entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

ENTRYPOINT ["entrypoint.sh"]

CMD ["npm", "run", "start"]

