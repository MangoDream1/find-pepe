import bodyParser from "body-parser";
import express from "express";
import helmet from "helmet";
import {
  PORT,
  DATA_DIR,
  BODY_LIMIT,
  PARAMETER_LIMIT,
  NODE_ENV,
  PUBLIC_SERVE_LOCATION,
} from "./constants";
import http from "http";
import path from "path";
import { router } from "./route";
import morgan from "morgan";

const config = () => {
  const app = express();

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

  app.use(helmet());
  app.use(
    PUBLIC_SERVE_LOCATION,
    express.static(path.join(__dirname, DATA_DIR))
  );

  app.use(router);

  return http.createServer(app);
};

const server = config();

const shutdown = async (): Promise<void> => {
  console.info(`Stopping server`);

  server.close((error) => {
    if (error) throw error;
    console.log("Stopped server");
    process.exit(0);
  });
};

const startup = async () => {
  process.on("unhandledRejection", (err) => {
    console.error(`unhandledRejection ${err}`);
    throw err;
  });

  process.on("uncaughtException", (err) => {
    console.error(`uncaughtException ${err}`);
    throw err;
  });

  process.on("SIGTERM", shutdown);
  process.on("SIGINT", shutdown);

  server.listen(PORT, "0.0.0.0", async () => {
    console.info(`Service listening on port ${PORT} [${NODE_ENV} mode]`);
  });
};

startup();
