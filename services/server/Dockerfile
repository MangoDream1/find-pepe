FROM node:18 as BUILD

WORKDIR /usr/src/app

COPY package.json ./
COPY yarn.lock ./

RUN yarn install

COPY . .

RUN yarn tsc
RUN ln -s ../public build/public

FROM node:18

WORKDIR /usr/src/app
COPY --from=BUILD /usr/src/app/build ./build
COPY --from=BUILD /usr/src/app/node_modules ./node_modules

EXPOSE 8080
CMD [ "node", "build/index.js" ]