import express from "express";
import { Core } from "../core.js";
import { DB } from "../db.js";
import { Controller } from "./controller.js";
import * as Validator from "./validator.js";

export const newRouter = (deps: { core: Core; db: DB }) => {
  const router = express.Router();

  const controller = new Controller(deps);

  router.get("/boards", controller.getBoards);
  router.get(
    "/boards/:category",
    Validator.getBoardsByCategoryValidator,
    controller.getBoardsByCategoryFromParam
  );
  router.get(
    "/",
    Validator.retrieveImagesByQueryRequest,
    controller.retrieveImagesByQuery
  );

  return router;
};
