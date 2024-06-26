import { useQuery } from "react-query";
import { API_URL } from "./constants";
import { Category } from "./types";

const GET_BOARDS_URL = () => new URL(`${API_URL.toString()}/boards`);
const GET_BOARDS_BY_CATEGORY_URL = (category: Category) =>
  new URL(`${GET_BOARDS_URL()}/${category}`);
const GET_IMAGE_PATHS_URL = () => new URL(API_URL);
const GET_IMAGE_BY_PATH_URL = (path: string) =>
  new URL(`${API_URL.toString()}${path}`);

function _handleNon200(response: Response) {
  throw new Error(
    `Failed to GET ${response.url} ${response.status} with message ${response.body}`
  );
}

export function useGetBoardsByCategory(category?: Category) {
  const queryFn = async () => {
    let url;
    if (category === undefined) {
      url = GET_BOARDS_URL();
    } else {
      url = GET_BOARDS_BY_CATEGORY_URL(category);
    }

    const response = await fetch(url);

    if (response.status !== 200) {
      _handleNon200(response);
    }
    return response.json();
  };

  return useQuery<string[]>({ queryKey: ["boards", category], queryFn });
}

export function useGetImagePaths(searchParams?: {
  category?: Category;
  board?: string;
  offset?: number;
}) {
  const queryFn = async () => {
    const url = GET_IMAGE_PATHS_URL();

    if (searchParams) {
      for (const [key, value] of Object.entries(searchParams)) {
        if (value === undefined) continue;

        url.searchParams.append(key, String(value));
      }
    }
    const response = await fetch(url);

    if (response.status !== 200) {
      _handleNon200(response);
    }
    return response.json();
  };

  return useQuery<string[]>({
    queryKey: [
      "imagePaths",
      searchParams?.category,
      searchParams?.board,
      searchParams?.offset,
    ],
    queryFn,
  });
}

export function transformPathToImagePath(path: string) {
  return GET_IMAGE_BY_PATH_URL(path).toString();
}
