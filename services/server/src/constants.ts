export const PORT = Number(process.env.PORT) || 5000;
export const DATA_DIR = process.env.DATA_DIR || "public/image";

export const BODY_LIMIT = process.env.BODY_LIMIT || "100mb";
export const PARAMETER_LIMIT = Number(process.env.PARAMETER_LIMIT) || 100;
export const NODE_ENV = process.env.NODE_ENV || "development";

export const IMAGE_CATEGORIES = ["maybe", "non-pepe", "pepe"] as const;

export const PUBLIC_SERVE_LOCATION = "/images";
