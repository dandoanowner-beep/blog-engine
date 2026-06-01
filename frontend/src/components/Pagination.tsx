interface Props {
  page: number
  total: number
  perPage: number
  onPageChange: (page: number) => void
}

export default function Pagination({ page, total, perPage, onPageChange }: Props) {
  const totalPages = Math.ceil(total / perPage)
  if (totalPages <= 1) return null

  const pages = Array.from({ length: totalPages }, (_, i) => i + 1)

  return (
    <nav className="flex justify-center gap-1 mt-6" aria-label="Pagination">
      <button
        onClick={() => onPageChange(page - 1)}
        disabled={page === 1}
        className="px-3 py-1 rounded border text-sm disabled:opacity-40 hover:bg-gray-50"
        aria-label="Previous page"
      >
        ‹
      </button>

      {pages.map((p) => (
        <button
          key={p}
          onClick={() => onPageChange(p)}
          aria-current={p === page ? 'page' : undefined}
          className={`px-3 py-1 rounded border text-sm ${
            p === page ? 'bg-blue-600 text-white border-blue-600' : 'hover:bg-gray-50'
          }`}
        >
          {p}
        </button>
      ))}

      <button
        onClick={() => onPageChange(page + 1)}
        disabled={page === totalPages}
        className="px-3 py-1 rounded border text-sm disabled:opacity-40 hover:bg-gray-50"
        aria-label="Next page"
      >
        ›
      </button>
    </nav>
  )
}
