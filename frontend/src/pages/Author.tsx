// OQ-005 (2026-06-11): the Author page is STATIC, to be built from the owner's
// design instructions (pending — see docs/OPEN_QUESTIONS.md). This is a minimal
// interim placeholder; no editing UI exists here by the owner's decision.
export default function Author() {
  return (
    <div className="max-w-3xl mx-auto">
      <h1 className="text-2xl font-bold mb-6">Author</h1>
      <div className="text-center py-12 text-gray-400">
        The author's story is coming soon.
      </div>
    </div>
  )
}
