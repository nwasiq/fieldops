interface PagerProps {
  page: number;
  pageSize: number;
  totalPages: number;
  total: number;
  onPage: (page: number) => void;
}

export function Pager({ page, pageSize, totalPages, total, onPage }: PagerProps) {
  const lastPage = Math.max(totalPages, 1);
  return (
    <nav className="pager" aria-label="Pagination">
      <button type="button" className="btn" disabled={page <= 1} onClick={() => onPage(page - 1)}>
        Previous
      </button>
      <span className="pager-status">
        Page {page} of {lastPage} · {pageSize} per page · {total} {total === 1 ? 'visit' : 'visits'}
      </span>
      <button type="button" className="btn" disabled={page >= lastPage} onClick={() => onPage(page + 1)}>
        Next
      </button>
    </nav>
  );
}
