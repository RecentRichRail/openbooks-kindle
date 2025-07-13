import {
  ColorScheme,
  ColorSchemeProvider,
  createEmotionCache,
  createStyles,
  MantineProvider
} from "@mantine/core";
import { useLocalStorage } from "@mantine/hooks";
import { NotificationsProvider } from "@mantine/notifications";
import NotificationDrawer from "./components/drawer/NotificationDrawer";
import SearchPage from "./pages/SearchPage";

const emotionCache = createEmotionCache({ key: "openbooks" });

const useStyles = createStyles((theme) => ({
  wrapper: {
    boxSizing: "border-box",
    width: "100vw",
    height: "100vh",
    margin: 0,
    padding: 0,
    backgroundColor:
      theme.colorScheme === "dark"
        ? theme.colors.dark[8]
        : theme.colors.gray[0],
    overflow: "hidden"
  }
}));

export default function App() {
  console.log("App component rendering...");
  const { classes } = useStyles();
  const [colorScheme, setColorScheme] = useLocalStorage({
    key: "color-scheme",
    defaultValue: "light" as ColorScheme,
    getInitialValueInEffect: true
  });

  console.log("App component state initialized, returning JSX...");

  return (
    <ColorSchemeProvider
      colorScheme={colorScheme}
      toggleColorScheme={() =>
        setColorScheme((color) => (color === "dark" ? "light" : "dark"))
      }>
      <MantineProvider
        emotionCache={emotionCache}
        withGlobalStyles
        withNormalizeCSS
        theme={{
          colorScheme,
          activeStyles: { transform: "none" },
          primaryColor: "brand",
          primaryShade: { light: 4, dark: 2 },
          colors: {
            brand: [
              "#e0ecff",
              "#b0c6ff",
              "#7e9fff",
              "#4c79ff",
              "#3366ff",
              "#0039e6",
              "#002db4",
              "#002082",
              "#001351",
              "#000621"
            ]
          },
          components: {
            ActionIcon: {
              defaultProps: {
                radius: "md",
                color: "brand"
              }
            }
          }
        }}>
        <NotificationsProvider position="top-center">
          <div className={classes.wrapper}>
            <SearchPage />
            <NotificationDrawer />
          </div>
        </NotificationsProvider>
      </MantineProvider>
    </ColorSchemeProvider>
  );
}
