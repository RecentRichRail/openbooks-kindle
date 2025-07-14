import {
  ActionIcon,
  Badge,
  Button,
  Center,
  createStyles,
  Group,
  Image,
  MediaQuery,
  Modal,
  Stack,
  Text,
  TextInput,
  Title,
  Tooltip
} from "@mantine/core";
import { useDisclosure } from "@mantine/hooks";
import { BellSimple, EnvelopeSimple, MagnifyingGlass, Warning, WifiHigh, WifiSlash } from "phosphor-react";
import { FormEvent, useEffect, useMemo, useState } from "react";
import image from "../assets/reading.svg";
import BookGrid from "../components/BookGrid";
import ErrorTable from "../components/tables/ErrorTable";
import { MessageType } from "../state/messages";
import { toggleDrawer } from "../state/notificationSlice";
import { sendMessage, sendSearch } from "../state/stateSlice";
import { useAppDispatch, useAppSelector } from "../state/store";

const useStyles = createStyles(
  (theme, { errorMode }: { errorMode: boolean }) => ({
    container: {
      width: "100vw",
      height: "100vh",
      margin: 0,
      padding: 0,
      position: "relative",
      display: "flex",
      flexDirection: "column"
    },
    header: {
      display: "flex",
      justifyContent: "space-between",
      alignItems: "center",
      padding: theme.spacing.md,
      borderBottom: `1px solid ${theme.colorScheme === "dark" ? theme.colors.dark[4] : theme.colors.gray[3]}`,
      backgroundColor: theme.colorScheme === "dark" ? theme.colors.dark[7] : theme.white,
      minHeight: 60
    },
    content: {
      flex: 1,
      padding: theme.spacing.md,
      overflow: "auto"
    },
    searchSection: {
      marginBottom: theme.spacing.md
    },
    wFull: {
      width: "100%"
    },
    errorToggle: {
      "alignSelf": "start",
      "height": "24px",
      "marginBottom": theme.spacing.xs,
      "fontWeight": 500,
      "color":
        theme.colorScheme === "dark"
          ? errorMode
            ? theme.colors.dark[8]
            : theme.colors.dark[2]
          : errorMode
          ? theme.colors.white
          : theme.colors.dark[3],
      "&:hover": {
        backgroundColor:
          theme.colorScheme === "dark"
            ? errorMode
              ? theme.colors.brand[3]
              : theme.colors.dark[7]
            : errorMode
            ? theme.colors.brand[5]
            : theme.colors.gray[1]
      }
    }
  })
);

export default function SearchPage() {
  const dispatch = useAppDispatch();
  const activeItem = useAppSelector((store) => store.state.activeItem);
  const isConnected = useAppSelector((store) => store.state.isConnected);
  const username = useAppSelector((store) => store.state.username);
  const { notifications } = useAppSelector((store) => store.notifications);

  const [searchQuery, setSearchQuery] = useState("");
  const [showErrors, setShowErrors] = useState(false);
  const [testEmailOpened, { open: openTestEmail, close: closeTestEmail }] = useDisclosure(false);
  const [testEmailAddress, setTestEmailAddress] = useState("");

  const hasErrors = (activeItem?.errors ?? []).length > 0;
  const errorMode = showErrors && activeItem;
  const validInput = errorMode
    ? searchQuery.startsWith("!")
    : searchQuery !== "";

  const { classes, theme } = useStyles({ errorMode: !!errorMode });

  const notificationCount = notifications.length;

  useEffect(() => {
    setShowErrors(false);
  }, [activeItem]);

  const searchHandler = (event: FormEvent) => {
    event.preventDefault();

    if (errorMode) {
      dispatch(
        sendMessage({
          type: MessageType.DOWNLOAD,
          payload: { book: searchQuery }
        })
      );
    } else {
      dispatch(sendSearch(searchQuery));
    }

    setSearchQuery("");
  };

  const handleTestEmail = () => {
    if (testEmailAddress.trim()) {
      dispatch(
        sendMessage({
          type: MessageType.SEND_TO_KINDLE,
          payload: {
            email: testEmailAddress.trim(),
            title: "SMTP Test",
            author: "OpenBooks",
            bookIdentifier: "test-smtp-email"
          }
        })
      );
      closeTestEmail();
      setTestEmailAddress("");
    }
  };

  const bookGrid = useMemo(
    () => <BookGrid books={activeItem?.results ?? []} />,
    [activeItem?.results]
  );

  const errorTable = useMemo(
    () => (
      <ErrorTable
        errors={activeItem?.errors ?? []}
        setSearchQuery={setSearchQuery}
      />
    ),
    [activeItem?.errors]
  );

  return (
    <div className={classes.container}>
      {/* Header with notification button */}
      <div className={classes.header}>
        <Title order={2} size="h3">OpenBooks</Title>
        <Group spacing="sm">
          <Tooltip label="Test SMTP Email">
            <ActionIcon
              variant="subtle"
              size="lg"
              onClick={openTestEmail}
              color="blue">
              <EnvelopeSimple size={20} />
            </ActionIcon>
          </Tooltip>
          <Tooltip label={isConnected ? `Connected to IRC as ${username || 'Unknown'}` : "Disconnected from IRC"}>
            <ActionIcon variant="subtle" size="lg">
              {isConnected ? (
                <WifiHigh size={20} color="green" />
              ) : (
                <WifiSlash size={20} color="red" />
              )}
            </ActionIcon>
          </Tooltip>
          <Tooltip label="Notifications">
            <ActionIcon
              variant="subtle"
              size="lg"
              onClick={() => dispatch(toggleDrawer())}>
              <BellSimple size={20} />
              {notificationCount > 0 && (
                <Badge
                  size="xs"
                  variant="filled"
                  color="red"
                  style={{
                    position: "absolute",
                    top: -2,
                    right: -2,
                    pointerEvents: "none"
                  }}>
                  {notificationCount}
                </Badge>
              )}
            </ActionIcon>
          </Tooltip>
        </Group>
      </div>

      {/* Main Content */}
      <div className={classes.content}>
        <div className={classes.searchSection}>
          <form className={classes.wFull} onSubmit={(e) => searchHandler(e)}>
            <Group
              noWrap
              spacing="md"
              sx={(theme) => ({ marginBottom: theme.spacing.sm })}>
              <TextInput
                className={classes.wFull}
                variant="filled"
                disabled={activeItem !== null && !activeItem.results}
                value={searchQuery}
                onChange={(e: any) => setSearchQuery(e.target.value)}
                placeholder={
                  errorMode ? "Download a book manually." : "Search for a book."
                }
                radius="md"
                type="search"
                icon={<MagnifyingGlass weight="bold" size={22} />}
                required
              />

              <Button
                type="submit"
                color={theme.colorScheme === "dark" ? "brand.2" : "brand"}
                disabled={!validInput}
                radius="md"
                variant={validInput ? "gradient" : "default"}
                gradient={{ from: "brand.4", to: "brand.3" }}>
                {errorMode ? "Download" : "Search"}
              </Button>
            </Group>
          </form>

          {hasErrors && (
            <Button
              className={classes.errorToggle}
              variant={errorMode ? "filled" : "subtle"}
              onClick={() => setShowErrors((show) => !show)}
              leftIcon={<Warning size={18} />}
              size="xs">
              {activeItem?.errors?.length} Parsing{" "}
              {activeItem?.errors?.length === 1 ? "Error" : "Errors"}
            </Button>
          )}
        </div>

        {!activeItem ? (
          <Center style={{ height: "calc(100vh - 200px)", width: "100%" }}>
            <Stack align="center" spacing="lg">
              <Title weight="normal" align="center" size="h3">
                Search a book to get started.
              </Title>
              <MediaQuery smallerThan="md" styles={{ display: "none" }}>
                <Image
                  width={400}
                  fit="contain"
                  src={image}
                  alt="person reading"
                />
              </MediaQuery>
              <MediaQuery largerThan="md" styles={{ display: "none" }}>
                <Image
                  width={250}
                  fit="contain"
                  src={image}
                  alt="person reading"
                />
              </MediaQuery>
            </Stack>
          </Center>
        ) : errorMode ? (
          errorTable
        ) : (
          bookGrid
        )}
      </div>

      {/* Test Email Modal */}
      <Modal
        opened={testEmailOpened}
        onClose={closeTestEmail}
        title="Test SMTP Email"
        size="md">
        <Stack spacing="md">
          <Text size="sm" color="dimmed">
            Send a test email to verify your SMTP configuration is working correctly.
          </Text>
          <TextInput
            label="Email Address"
            placeholder="test@example.com"
            value={testEmailAddress}
            onChange={(event) => setTestEmailAddress(event.currentTarget.value)}
            required
          />
          <Group position="right" spacing="sm">
            <Button variant="subtle" onClick={closeTestEmail}>
              Cancel
            </Button>
            <Button 
              onClick={handleTestEmail}
              disabled={!testEmailAddress.trim()}>
              Send Test Email
            </Button>
          </Group>
        </Stack>
      </Modal>
    </div>
  );
}
