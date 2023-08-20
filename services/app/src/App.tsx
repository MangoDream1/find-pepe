import { useGetImagePaths, transformPathToImagePath } from "./hooks";
import Box from "@mui/material/Box";
import { ImageList, ImageListItem } from "@mui/material";

function App() {
  const response = useGetImagePaths();

  if (response.error) return <>An error has occurred</>;
  if (response.isLoading || !response.data) return <>Loading...</>;

  return (
    <Box>
      <ImageList variant="masonry" cols={5}>
        {response.data.slice(0, 200).map((s, index) => (
          <ImageListItem key={index}>
            <img
              src={transformPathToImagePath(s)}
              loading="lazy"
              style={{ width: "100%" }}
            />
          </ImageListItem>
        ))}
      </ImageList>
    </Box>
  );
}

export default App;
