import express from "express";
import { DB } from "../db";
import { Category } from "../types";
import { Core } from "../core";

export class Controller {
  core: Core;
  db: DB;

  constructor(deps: { core: Core; db: DB }) {
    this.core = deps.core;
    this.db = deps.db;
  }

  async getBoards(req: express.Request, res: express.Response) {
    const boards = await this.db.getUniqueBoards();
    res.status(200).send(boards);
  }

  async getBoardsByCategoryFromParam(
    req: express.Request,
    res: express.Response
  ) {
    const boards = await this.db.getUniqueBoardsByCategory(
      req.params.category as Category
    );
    res.status(200).send(boards);
  }

  // TODO: add skip / limit
  async retrieveImagesByQuery(req: express.Request, res: express.Response) {
    const category = req.query.category as Category | undefined;
    const board = req.query.board as string | undefined;

    const selection = {
      categories: [] as Category[],
      boards: [] as string[],
    };
    if (category) {
      selection.categories = [category];
    }
    if (board) {
      selection.boards = [board];
    }

    const images: string[] = [];
    const imageGen = this.db.getImagesBySelection(selection);
    for await (const img of imageGen) {
      images.push(this.core.removeDataDirFromPath(img));
    }
    imageGen.return(undefined);

    res.status(200).send(Array.from(images));
  }
}
