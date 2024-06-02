import path from "path";

export const PORT = Number(process.env.PORT) || 5000;
export const DATA_DIR = process.env.DATA_DIR || "public/image";
export const DATA_PATH = path.join(path.resolve(), DATA_DIR);

export const BODY_LIMIT = process.env.BODY_LIMIT || "100mb";
export const PARAMETER_LIMIT = Number(process.env.PARAMETER_LIMIT) || 100;
export const NODE_ENV = process.env.NODE_ENV || "development";

export const IMAGE_CATEGORIES = ["maybe", "non-pepe", "pepe"] as const;

export const PUBLIC_SERVE_LOCATION = "/images";

export const IMAGE_RETURN_LIMIT = Number(process.env.IMAGE_RETURN_LIMIT) || 50;
