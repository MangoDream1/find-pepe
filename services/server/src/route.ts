import express from "express";
import * as Core from "./core";
import { IMAGE_CATEGORIES } from "./constants";
import { Category } from "./types";

export const router = express.Router();

router.get("/boards", (req, res) => {
  const boards = Array.from(
    Core.mergeUniqueStringGenerators(
      IMAGE_CATEGORIES.map((c: Category) => Core.getBoards(c))
    )
  );

  res.status(200).send(boards);
});

const boardsCategoryValidator = (
  req: express.Request,
  res: express.Response,
  next: express.NextFunction
) => {
  const category = req.params.category as Category;
  if (!IMAGE_CATEGORIES.includes(category)) {
    res
      .status(400)
      .send(`Category should be one of: [${IMAGE_CATEGORIES.join(", ")}]`);
    return;
  }

  next();
};

router.get("/boards/:category", boardsCategoryValidator, (req, res) => {
  const category = req.params.category as Category;

  res.status(200).send(Array.from(Core.getBoards(category)));
});

const rootValidator = (
  req: express.Request,
  res: express.Response,
  next: express.NextFunction
) => {
  const category = req.query.category as Category;
  if (category !== undefined && !IMAGE_CATEGORIES.includes(category)) {
    res
      .status(400)
      .send(`Category should be one of: [${IMAGE_CATEGORIES.join(", ")}]`);
    return;
  }

  next();
};

function* _retrieveImagesByBoardsAndCategories(
  boards: Iterable<string>,
  categories: readonly Category[]
): Generator<string, void> {
  const imagePaths: Generator<string, void>[] = [];
  for (const b of boards) {
    for (const c of categories) {
      imagePaths.push(Core.retrieveImagePaths(c, b));
    }
  }

  for (const path of Core.mergeStringGenerators(imagePaths)) {
    yield Core.removeDataDirFromPath(path);
  }
}

router.get("/", rootValidator, (req, res) => {
  const category = req.query.category as Category | undefined;
  const board = req.query.board as string | undefined;

  if (category && board) {
    res.status(200).send(Array.from(Core.retrieveImagePaths(category, board)));
    return;
  }

  if (category) {
    const boards = Core.getBoards(category);
    res
      .status(200)
      .send(
        Array.from(_retrieveImagesByBoardsAndCategories(boards, [category]))
      );
    return;
  }

  if (board) {
    res
      .status(200)
      .send(
        Array.from(
          _retrieveImagesByBoardsAndCategories([board], IMAGE_CATEGORIES)
        )
      );
    return;
  }

  const boards = Core.mergeUniqueStringGenerators(
    IMAGE_CATEGORIES.map((c: Category) => Core.getBoards(c))
  );

  res
    .status(200)
    .send(
      Array.from(_retrieveImagesByBoardsAndCategories(boards, IMAGE_CATEGORIES))
    );
});
