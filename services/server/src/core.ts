import fs from "fs";
import path from "path";
import { DATA_DIR, PUBLIC_SERVE_LOCATION } from "./constants";
import { Category } from "./types";

function _createDataDirPath() {
  return path.join(__dirname, "..", DATA_DIR);
}

function* getServerPaths(category: Category): Generator<string, void> {
  const categoryDir = path.join(_createDataDirPath(), category);
  const servers = fs.readdirSync(categoryDir);

  for (const server of servers) {
    const r = path.join(categoryDir, server);
    yield r;
  }
}

function* getBoardPaths(serverPath: string): Generator<string, void> {
  const boardDirectories = fs.readdirSync(serverPath);
  for (const board of boardDirectories) {
    yield board;
  }
}

export function* getBoards(category: Category): Generator<string, void> {
  const serverPaths = getServerPaths(category);
  for (const serverPath of serverPaths) {
    const boardPaths = getBoardPaths(serverPath);
    yield* boardPaths;
  }
}

export function* mergeUniqueStringGenerators(
  boards: Generator<string, void>[]
): Generator<string, void> {
  const unique: string[] = [];

  for (const gen of boards) {
    for (const board of gen) {
      if (unique.includes(board)) continue;
      unique.push(board);
      yield board;
    }
  }
}

export function* mergeStringGenerators(
  boards: Generator<string, void>[]
): Generator<string, void> {
  for (const gen of boards) {
    yield* gen;
  }
}

export function* retrieveImagePaths(category: Category, board: string) {
  const serverPaths = getServerPaths(category);

  for (const serverPath of serverPaths) {
    const imageDir = path.join(serverPath, board);

    if (!fs.existsSync(imageDir)) return;

    const images = fs.readdirSync(imageDir);
    for (const image of images) {
      const imagePath = path.join(imageDir, image);
      yield imagePath;
    }
  }
}

export function removeDataDirFromPath(p: string) {
  const dirPath = _createDataDirPath();
  return path.join(PUBLIC_SERVE_LOCATION, p.replace(dirPath, ""));
}
