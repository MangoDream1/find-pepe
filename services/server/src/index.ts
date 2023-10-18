import bodyParser from "body-parser";
import cors from "cors";
import express from "express";
import helmet from "helmet";
import http from "http";
import morgan from "morgan";
import path from "path";
import { Sequelize } from "sequelize";
import {
  BODY_LIMIT,
  DATA_DIR,
  NODE_ENV,
  PARAMETER_LIMIT,
  PORT,
  PUBLIC_SERVE_LOCATION,
} from "./constants";
import { Core } from "./core";
import { DB } from "./db";
import { newRouter } from "./route";
import sequelize from "sequelize";

const config = () => {
  const app = express();

  // enable cors for development
  if (NODE_ENV === "development") {
    app.use(cors({ credentials: true }));
  }

  if (NODE_ENV === "production") {
    app.use(helmet());
  }

  app.use(
    bodyParser.json({
      limit: BODY_LIMIT,
    })
  );
  app.use(
    bodyParser.urlencoded({
      extended: true,
      limit: BODY_LIMIT,
      parameterLimit: PARAMETER_LIMIT,
    })
  );
  app.use(morgan("common"));

  app.use(
    PUBLIC_SERVE_LOCATION,
    express.static(path.join(__dirname, DATA_DIR))
  );

  return app;
};

const shutdown =
  (deps: { sequelize: Sequelize; server: http.Server }) =>
  async (): Promise<void> => {
    const { server, sequelize } = deps;

    console.info(`Stopping server`);
    await sequelize.close();

    server.close((error) => {
      if (error) throw error;
      console.log("Stopped server");
      process.exit(0);
    });
  };

const startup = async () => {
  // TODO: put in env variables
  const sequelize = new Sequelize("postgres", "admin", "test", {
    dialect: "postgres",
    host: "postgresql",
    port: 5432,
  });

  await sequelize.authenticate();

  const app = config();

  const core = new Core();
  const db = new DB({ sequelize });

  app.use(newRouter({ core, db }));

  const server = http.createServer(app);

  process.on("unhandledRejection", (err) => {
    console.error(`unhandledRejection ${err}`);
    throw err;
  });

  process.on("uncaughtException", (err) => {
    console.error(`uncaughtException ${err}`);
    throw err;
  });

  process.on("SIGTERM", shutdown({ server, sequelize }));
  process.on("SIGINT", shutdown({ server, sequelize }));

  server.listen(PORT, "0.0.0.0", async () => {
    console.info(`Service listening on port ${PORT} [${NODE_ENV} mode]`);
  });
};

startup();
