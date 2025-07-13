import {
  Grid,
  SimpleGrid,
  Text,
  Stack,
  Group,
  TextInput,
  Select
} from "@mantine/core";
import { MagnifyingGlass } from "phosphor-react";
import { useState, useMemo } from "react";
import { BookDetail } from "../state/messages";
import BookCard from "./BookCard";

interface BookGridProps {
  books: BookDetail[];
}

export default function BookGrid({ books }: BookGridProps) {
  const [searchQuery, setSearchQuery] = useState("");
  const [formatFilter, setFormatFilter] = useState<string | null>(null);
  const [serverFilter, setServerFilter] = useState<string | null>(null);

  // Filter books based on search and filters
  const filteredBooks = useMemo(() => {
    return books.filter((book) => {
      const matchesSearch = 
        book.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
        book.author.toLowerCase().includes(searchQuery.toLowerCase());
      
      const matchesFormat = !formatFilter || book.format === formatFilter;
      const matchesServer = !serverFilter || book.server === serverFilter;
      
      return matchesSearch && matchesFormat && matchesServer;
    });
  }, [books, searchQuery, formatFilter, serverFilter]);

  // Get unique values for filters
  const formats = useMemo(() => {
    const uniqueFormats = Array.from(new Set(books.map(book => book.format).filter(Boolean)));
    return uniqueFormats.map(format => ({ value: format, label: format.toUpperCase() }));
  }, [books]);

  const servers = useMemo(() => {
    const uniqueServers = Array.from(new Set(books.map(book => book.server)));
    return uniqueServers.map(server => ({ value: server, label: server }));
  }, [books]);

  if (books.length === 0) {
    return (
      <Stack align="center" spacing="md" style={{ padding: "2rem" }}>
        <Text size="lg" color="dimmed">
          No books found
        </Text>
      </Stack>
    );
  }

  return (
    <Stack spacing="md">
      {/* Filters */}
      <Group spacing="md" align="end">
        <TextInput
          placeholder="Search books or authors..."
          icon={<MagnifyingGlass size={16} />}
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          style={{ flex: 1, minWidth: 200 }}
        />
        
        <Select
          placeholder="Format"
          data={formats}
          value={formatFilter}
          onChange={setFormatFilter}
          clearable
          style={{ minWidth: 120 }}
        />
        
        <Select
          placeholder="Server"
          data={servers}
          value={serverFilter}
          onChange={setServerFilter}
          clearable
          style={{ minWidth: 150 }}
        />
      </Group>

      {/* Results count */}
      <Text size="sm" color="dimmed">
        Showing {filteredBooks.length} of {books.length} books
      </Text>

      {/* Book Grid */}
      <SimpleGrid
        cols={1}
        spacing="md"
        breakpoints={[
          { minWidth: 'sm', cols: 1 },
          { minWidth: 'md', cols: 2 },
          { minWidth: 'lg', cols: 3 },
          { minWidth: 1400, cols: 4 }
        ]}>
        {filteredBooks.map((book, index) => (
          <BookCard key={`${book.title}-${book.author}-${index}`} book={book} />
        ))}
      </SimpleGrid>

      {filteredBooks.length === 0 && searchQuery && (
        <Stack align="center" spacing="md" style={{ padding: "2rem" }}>
          <Text size="lg" color="dimmed">
            No books match your search
          </Text>
          <Text size="sm" color="dimmed">
            Try adjusting your search terms or filters
          </Text>
        </Stack>
      )}
    </Stack>
  );
}
