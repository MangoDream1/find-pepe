FROM node:18 as BUILD

WORKDIR /usr/src/app

COPY package.json ./
COPY yarn.lock ./

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