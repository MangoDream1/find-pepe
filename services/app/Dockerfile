FROM node:18 as BUILD

WORKDIR /usr/src/app

RUN yarn config set enableNetwork false

COPY package.json ./
COPY yarn.lock ./
COPY .yarn ./
COPY .pnp* ./

RUN yarn install

COPY . .

RUN yarn build

FROM nginx:1.16-alpine

WORKDIR /usr/share/nginx/html

RUN sed -i '/location \//a try_files $uri $uri/ /index.html;' /etc/nginx/conf.d/default.conf

COPY --from=BUILD /usr/src/app/dist .

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
