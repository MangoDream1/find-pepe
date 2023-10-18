import express from "express";
import { Controller } from "./controller";
import { Validator } from "./validator";
import { DB } from "../db";
import { Core } from "../core";

export const newRouter = (deps: { core: Core; db: DB }) => {
  const router = express.Router();

  const controller = new Controller(deps);
  const validator = new Validator();

  router.get("/boards", controller.getBoards);
  router.get(
    "/boards/:category",
    validator.paramCategory,
    controller.getBoardsByCategoryFromParam
  );
  router.get("/", validator.queryCategory, controller.retrieveImagesByQuery);

  return router;
};
