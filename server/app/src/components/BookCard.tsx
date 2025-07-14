import {
  Card,
  Group,
  Text,
  Badge,
  Button,
  Image,
  Stack,
  Indicator,
  Tooltip,
  createStyles
} from "@mantine/core";
import { useState } from "react";
import { BookDetail } from "../state/messages";
import SendToKindle from "./SendToKindle";
import { useGetServersQuery } from "../state/api";

const useStyles = createStyles((theme) => ({
  card: {
    backgroundColor: theme.colorScheme === "dark" ? theme.colors.dark[7] : theme.white,
    borderRadius: theme.radius.md,
    border: `1px solid ${theme.colorScheme === "dark" ? theme.colors.dark[5] : theme.colors.gray[3]}`,
    padding: theme.spacing.md,
    boxShadow: theme.shadows.sm,
    transition: "transform 0.2s ease, box-shadow 0.2s ease",
    "&:hover": {
      transform: "translateY(-2px)",
      boxShadow: theme.shadows.md
    }
  },
  coverContainer: {
    minWidth: 60,
    width: 60,
    height: 80,
    display: "flex",
    alignItems: "center",
    justifyContent: "center",
    backgroundColor: theme.colorScheme === "dark" ? theme.colors.dark[6] : theme.colors.gray[1],
    borderRadius: theme.radius.sm,
    border: `1px solid ${theme.colorScheme === "dark" ? theme.colors.dark[5] : theme.colors.gray[3]}`
  },
  cover: {
    width: "100%",
    height: "100%",
    objectFit: "cover",
    borderRadius: theme.radius.sm
  },
  placeholder: {
    backgroundColor: theme.colorScheme === "dark" ? theme.colors.dark[5] : theme.colors.gray[2],
    border: `1px solid ${theme.colorScheme === "dark" ? theme.colors.dark[4] : theme.colors.gray[3]}`,
    borderRadius: 4,
    display: "flex",
    alignItems: "center",
    justifyContent: "center",
    fontSize: 10,
    color: theme.colorScheme === "dark" ? theme.colors.dark[2] : theme.colors.gray[6]
  },
  contentSection: {
    flex: 1,
    minWidth: 0
  },
  title: {
    fontWeight: 600,
    lineHeight: 1.3,
    marginBottom: 4
  },
  author: {
    color: theme.colorScheme === "dark" ? theme.colors.gray[4] : theme.colors.gray[6],
    marginBottom: 8
  },
  metadata: {
    gap: 8
  },
  serverBadge: {
    fontSize: 11
  }
}));

interface BookCardProps {
  book: BookDetail;
}

export default function BookCard({ book }: BookCardProps) {
  const { classes } = useStyles();
  const { data: servers } = useGetServersQuery(null);
  const isServerOnline = servers?.includes(book.server) ?? false;

  const BookCover = ({ title, author }: { title: string; author: string }) => {
    const [imageLoaded, setImageLoaded] = useState(false);
    const [imageError, setImageError] = useState(false);
    
    const coverUrl = `https://covers.openlibrary.org/b/title/${encodeURIComponent(title)}-M.jpg`;
    
    return (
      <div className={classes.coverContainer}>
        {!imageError ? (
          <img
            src={coverUrl}
            alt={`Cover for ${title}`}
            className={classes.cover}
            onLoad={() => setImageLoaded(true)}
            onError={() => setImageError(true)}
            style={{ display: imageLoaded ? "block" : "none" }}
          />
        ) : null}
        {(!imageLoaded || imageError) && (
          <div className={classes.placeholder}>
            No Cover
          </div>
        )}
      </div>
    );
  };

  return (
    <Card className={classes.card}>
      <Group spacing="md" align="flex-start" noWrap>
        <BookCover title={book.title} author={book.author} />
        
        <div className={classes.contentSection}>
          <Text className={classes.title} lineClamp={2}>
            {book.title}
          </Text>
          
          <Text size="sm" className={classes.author} lineClamp={1}>
            by {book.author}
          </Text>
          
          <Group className={classes.metadata} spacing="xs" mb="sm">
            <Tooltip label={isServerOnline ? "Server Online" : "Server Offline"}>
              <Badge
                size="xs"
                variant={isServerOnline ? "filled" : "outline"}
                color={isServerOnline ? "green" : "gray"}
                className={classes.serverBadge}>
                {book.server}
              </Badge>
            </Tooltip>
            
            {book.format && (
              <Badge size="xs" variant="light" color="blue">
                {book.format.toUpperCase()}
              </Badge>
            )}
            
            {book.size && (
              <Badge size="xs" variant="light" color="gray">
                {book.size}
              </Badge>
            )}
          </Group>
          
          <Group position="right">
            <SendToKindle 
              book={book.full}
              title={book.title}
              author={book.author}
            />
          </Group>
        </div>
      </Group>
    </Card>
  );
}
