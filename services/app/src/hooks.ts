import { useQuery } from "react-query";
import { API_URL } from "./constants";
import { Category } from "./types";

const BASE_URL = new URL("/", API_URL);

const GET_BOARDS_URL = () => new URL("/boards", BASE_URL);
const GET_BOARDS_BY_CATEGORY_URL = (category: Category) =>
  new URL(category, GET_BOARDS_URL());
const GET_IMAGE_PATHS_URL = () => new URL("/", BASE_URL);
const GET_IMAGE_BY_PATH_URL = (path: string) => new URL(path, BASE_URL);

function _handleNon200(response: Response) {
  throw new Error(
    `Failed to GET ${response.url} ${response.status} with message ${response.body}`
  );
}

export function useGetBoards() {
  const queryFn = async () => {
    const response = await fetch(GET_BOARDS_URL());

    if (response.status !== 200) {
      _handleNon200(response);
    }
    return response.json();
  };

  return useQuery<string[]>({ queryKey: ["boards"], queryFn });
}

export function useGetBoardsByCategory(category: Category) {
  const queryFn = async () => {
    const url = GET_BOARDS_BY_CATEGORY_URL(category);
    const response = await fetch(url);

    if (response.status !== 200) {
      _handleNon200(response);
    }
    return response.json();
  };

  return useQuery<string[]>({ queryKey: ["boards"], queryFn });
}

export function useGetImagePaths(searchParams?: {
  category?: Category;
  board?: string;
}) {
  const queryFn = async () => {
    const url = GET_IMAGE_PATHS_URL();
    console.log(url);

    if (searchParams) {
      for (const [key, value] of Object.entries(searchParams)) {
        url.searchParams.append(key, value);
      }
    }
    const response = await fetch(url);

    if (response.status !== 200) {
      _handleNon200(response);
    }
    return response.json();
  };

  return useQuery<string[]>({ queryKey: ["imagePaths"], queryFn });
}

export function transformPathToImagePath(path: string) {
  console.log(path, GET_IMAGE_BY_PATH_URL(path).toString());
  return GET_IMAGE_BY_PATH_URL(path).toString();
}
