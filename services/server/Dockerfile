FROM node:18 as BUILD

WORKDIR /usr/src/app

RUN yarn config set enableNetwork false

COPY . .

RUN yarn install
RUN yarn build

FROM node:18

WORKDIR /usr/src/app
COPY --from=BUILD /usr/src/app/package.json ./package.json
COPY --from=BUILD /usr/src/app/build ./build

EXPOSE 5000
CMD [ "yarn", "start" ]