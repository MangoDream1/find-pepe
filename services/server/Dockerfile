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

FROM node:18

WORKDIR /usr/src/app
COPY --from=BUILD /usr/src/app/package.json ./package.json
COPY --from=BUILD /usr/src/app/build ./build
COPY --from=BUILD /usr/src/app/node_modules ./node_modules

EXPOSE 5000
CMD [ "yarn", "start" ]