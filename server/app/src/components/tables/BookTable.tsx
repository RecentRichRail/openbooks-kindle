import {
  Image,
  Indicator,
  Loader,
  ScrollArea,
  Table,
  Text,
  Tooltip
} from "@mantine/core";
import { useElementSize, useMergedRef } from "@mantine/hooks";
import {
  createColumnHelper,
  FilterFn,
  flexRender,
  getCoreRowModel,
  getFacetedRowModel,
  getFacetedUniqueValues,
  getFilteredRowModel,
  Row,
  useReactTable
} from "@tanstack/react-table";
import { useVirtualizer } from "@tanstack/react-virtual";
import { MagnifyingGlass, User, BookOpen } from "phosphor-react";
import { useMemo, useRef, useState, useEffect } from "react";
import { useGetServersQuery } from "../../state/api";
import { BookDetail } from "../../state/messages";
import SendToKindle from "../SendToKindle";
import { useAppDispatch } from "../../state/store";
import FacetFilter, {
  ServerFacetEntry,
  StandardFacetEntry
} from "./Filters/FacetFilter";
import { TextFilter } from "./Filters/TextFilter";
import { useTableStyles } from "./styles";

const columnHelper = createColumnHelper<BookDetail>();

const stringInArray: FilterFn<any> = (
  row,
  columnId: string,
  filterValue: string[] | undefined
) => {
  if (!filterValue || filterValue.length === 0) return true;

  return filterValue.includes(row.getValue<string>(columnId));
};

interface BookTableProps {
  books: BookDetail[];
}

export default function BookTable({ books }: BookTableProps) {
  const { classes, cx, theme } = useTableStyles();
  const { data: servers } = useGetServersQuery(null);

  const { ref: elementSizeRef, height, width } = useElementSize();
  const virtualizerRef = useRef();
  const mergedRef = useMergedRef(elementSizeRef, virtualizerRef);

  const columns = useMemo(() => {
    const cols = (cols: number) => (width / 12) * cols;
    return [
      // New Cover Column
      columnHelper.display({
        id: "cover",
        header: "Cover",
        size: cols(0.8),
        enableColumnFilter: false,
        cell: ({ row }) => (
          <BookCover 
            title={row.original.title} 
            author={row.original.author}
          />
        )
      }),
      columnHelper.accessor("server", {
        header: (props) => (
          <FacetFilter
            placeholder="Server"
            column={props.column}
            table={props.table}
            Entry={ServerFacetEntry}
          />
        ),
        cell: (props) => {
          const online = servers?.includes(props.getValue());
          return (
            <Text
              size={12}
              weight="normal"
              color="dark"
              style={{ marginLeft: 20 }}>
              <Tooltip
                position="top-start"
                label={online ? "Online" : "Offline"}>
                <Indicator
                  zIndex={0}
                  position="middle-start"
                  offset={-16}
                  size={6}
                  color={online ? "green.6" : "gray"}>
                  {props.getValue()}
                </Indicator>
              </Tooltip>
            </Text>
          );
        },
        size: cols(1),
        enableColumnFilter: true,
        filterFn: stringInArray
      }),
      // Enhanced Author Column
      columnHelper.accessor("author", {
        header: (props) => (
          <TextFilter
            icon={<User weight="bold" />}
            placeholder="Author"
            column={props.column}
            table={props.table}
          />
        ),
        cell: (props) => (
          <div style={{ padding: "4px 0" }}>
            <Text
              size="sm"
              weight={500}
              color="dark"
              lineClamp={2}
              style={{ lineHeight: 1.3 }}>
              {props.getValue()}
            </Text>
          </div>
        ),
        size: cols(2.2),
        enableColumnFilter: false
      }),
      columnHelper.accessor("title", {
        header: (props) => (
          <TextFilter
            icon={<MagnifyingGlass weight="bold" />}
            placeholder="Title"
            column={props.column}
            table={props.table}
          />
        ),
        minSize: 20,
        size: cols(5),
        enableColumnFilter: false
      }),
      columnHelper.accessor("format", {
        header: (props) => (
          <FacetFilter
            placeholder="Format"
            column={props.column}
            table={props.table}
            Entry={StandardFacetEntry}
          />
        ),
        size: cols(1),
        enableColumnFilter: false,
        filterFn: stringInArray
      }),
      columnHelper.accessor("size", {
        header: "Size",
        size: cols(1),
        enableColumnFilter: false
      }),
      columnHelper.display({
        header: "Send to Kindle",
        size: cols(1.2),
        enableColumnFilter: false,
        cell: ({ row }) => (
          <SendToKindle 
            book={row.original.full}
            title={row.original.title}
            author={row.original.author}
          />
        )
      })
    ];
  }, [width, servers]);

  const table = useReactTable({
    data: books,
    columns: columns,
    enableFilters: true,
    columnResizeMode: "onChange",
    getCoreRowModel: getCoreRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getFacetedRowModel: getFacetedRowModel(),
    getFacetedUniqueValues: getFacetedUniqueValues()
  });

  const { rows: tableRows } = table.getRowModel();

  const rowVirtualizer = useVirtualizer({
    count: tableRows.length,
    getScrollElement: () => virtualizerRef.current,
    estimateSize: () => 70, // Increased height for book covers
    overscan: 10
  });

  const virtualItems = rowVirtualizer.getVirtualItems();

  const paddingTop =
    virtualItems.length > 0 ? virtualItems?.[0]?.start || 0 : 0;
  const paddingBottom =
    virtualItems.length > 0
      ? rowVirtualizer.getTotalSize() -
        (virtualItems?.[virtualItems.length - 1]?.end || 0)
      : 0;

  return (
    <ScrollArea
      viewportRef={mergedRef}
      className={classes.container}
      type="hover"
      scrollbarSize={6}
      styles={{ thumb: { ["&::before"]: { minWidth: 4 } } }}
      offsetScrollbars={false}>
      <Table highlightOnHover verticalSpacing="sm" fontSize="xs">
        <thead className={classes.head}>
          {table.getHeaderGroups().map((headerGroup) => (
            <tr key={headerGroup.id}>
              {headerGroup.headers.map((header) => (
                <th
                  key={header.id}
                  className={classes.headerCell}
                  style={{
                    width: header.getSize()
                  }}>
                  {flexRender(
                    header.column.columnDef.header,
                    header.getContext()
                  )}
                  <div
                    onMouseDown={header.getResizeHandler()}
                    onTouchStart={header.getResizeHandler()}
                    className={cx(classes.resizer, {
                      ["isResizing"]: header.column.getIsResizing()
                    })}
                  />
                </th>
              ))}
            </tr>
          ))}
        </thead>
        <tbody>
          {paddingTop > 0 && (
            <tr>
              <td style={{ height: `${paddingTop}px` }} />
            </tr>
          )}
          {rowVirtualizer.getVirtualItems().map((virtualRow) => {
            const row = tableRows[
              virtualRow.index
            ] as unknown as Row<BookDetail>;
            return (
              <tr key={row.id} style={{ height: 70 }}>
                {row.getVisibleCells().map((cell) => {
                  return (
                    <td key={cell.id} style={{ verticalAlign: 'middle' }}>
                      <Text lineClamp={1} color="dark">
                        {flexRender(
                          cell.column.columnDef.cell,
                          cell.getContext()
                        )}
                      </Text>
                    </td>
                  );
                })}
              </tr>
            );
          })}
          {paddingBottom > 0 && (
            <tr>
              <td style={{ height: `${paddingBottom}px` }} />
            </tr>
          )}
        </tbody>
      </Table>
    </ScrollArea>
  );
}

// Component for book cover display
function BookCover({ title, author }: { title: string; author: string }) {
  const [coverUrl, setCoverUrl] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(false);

  // Fetch book cover from multiple sources
  const fetchBookCover = async (title: string, author: string) => {
    if (!title || loading) return;
    
    setLoading(true);
    setError(false);

    try {
      // Try Google Books API first
      const googleUrl = await fetchGoogleBooksCover(title, author);
      if (googleUrl) {
        setCoverUrl(googleUrl);
        setLoading(false);
        return;
      }

      // Fallback to Open Library API
      const openLibraryUrl = await fetchOpenLibraryCover(title, author);
      if (openLibraryUrl) {
        setCoverUrl(openLibraryUrl);
        setLoading(false);
        return;
      }

      // No cover found
      setError(true);
    } catch (err) {
      console.warn('Error fetching book cover:', err);
      setError(true);
    }
    
    setLoading(false);
  };

  // Google Books API integration
  const fetchGoogleBooksCover = async (title: string, author: string): Promise<string | null> => {
    try {
      const query = encodeURIComponent(`${title} ${author}`);
      const response = await fetch(
        `https://www.googleapis.com/books/v1/volumes?q=${query}&maxResults=1`
      );
      const data = await response.json();
      
      if (data.items && data.items[0]?.volumeInfo?.imageLinks?.thumbnail) {
        // Upgrade to higher quality if available
        let imageUrl = data.items[0].volumeInfo.imageLinks.thumbnail;
        if (data.items[0].volumeInfo.imageLinks.smallThumbnail) {
          imageUrl = data.items[0].volumeInfo.imageLinks.smallThumbnail.replace('&zoom=1', '&zoom=2');
        }
        return imageUrl.replace('http://', 'https://');
      }
    } catch (error) {
      console.warn('Google Books API error:', error);
    }
    return null;
  };

  // Open Library API integration
  const fetchOpenLibraryCover = async (title: string, author: string): Promise<string | null> => {
    try {
      const query = encodeURIComponent(`${title} ${author}`);
      const response = await fetch(
        `https://openlibrary.org/search.json?title=${encodeURIComponent(title)}&author=${encodeURIComponent(author)}&limit=1`
      );
      const data = await response.json();
      
      if (data.docs && data.docs[0]?.cover_i) {
        const coverId = data.docs[0].cover_i;
        return `https://covers.openlibrary.org/b/id/${coverId}-M.jpg`;
      }
    } catch (error) {
      console.warn('Open Library API error:', error);
    }
    return null;
  };

  // Fetch cover when component mounts or props change
  useEffect(() => {
    const timeoutId = setTimeout(() => {
      fetchBookCover(title, author);
    }, Math.random() * 1000); // Stagger requests to avoid rate limiting

    return () => clearTimeout(timeoutId);
  }, [title, author]);

  return (
    <div style={{ width: 40, height: 60, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
      {loading ? (
        <div 
          style={{ 
            width: 40, 
            height: 60, 
            backgroundColor: '#f1f3f4',
            border: '1px solid #e0e0e0',
            borderRadius: 4,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center'
          }}
        >
          <Loader size="xs" />
        </div>
      ) : coverUrl ? (
        <Image
          src={coverUrl}
          alt={`Cover for ${title} by ${author}`}
          width={40}
          height={60}
          fit="cover"
          radius="sm"
          withPlaceholder
          style={{
            border: '1px solid #e0e0e0',
            boxShadow: '0 2px 4px rgba(0,0,0,0.1)'
          }}
        />
      ) : (
        <div 
          style={{ 
            width: 40, 
            height: 60, 
            backgroundColor: '#f8f9fa',
            border: '2px dashed #dee2e6',
            borderRadius: 4,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            flexDirection: 'column'
          }}
        >
          <BookOpen size={16} color="#adb5bd" />
          <Text size={8} color="dimmed" align="center" style={{ marginTop: 2, lineHeight: 1 }}>
            No Cover
          </Text>
        </div>
      )}
    </div>
  );
}
