import { useState, useCallback, useEffect } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { useEditor, EditorContent } from '@tiptap/react'
import StarterKit from '@tiptap/starter-kit'
import Image from '@tiptap/extension-image'
import Link from '@tiptap/extension-link'
import CodeBlockLowlight from '@tiptap/extension-code-block-lowlight'
import { common, createLowlight } from 'lowlight'
import { blogsApi, type CreateBlogInput } from '../api/blogs'

const lowlight = createLowlight(common)

const PRIVACY_OPTIONS = [
  { value: 'public', label: 'Public' },
  { value: 'friend_only', label: 'Friends only' },
  { value: 'only_me', label: 'Only me' },
] as const

export default function Editor() {
  const { id } = useParams<{ id?: string }>()
  const navigate = useNavigate()
  const [title, setTitle] = useState('')
  const [privacy, setPrivacy] = useState<CreateBlogInput['privacy']>('public')
  const [tags, setTags] = useState('')
  const [thumbnail, setThumbnail] = useState('')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  const editor = useEditor({
    extensions: [
      StarterKit,
      Image,
      Link.configure({ openOnClick: false }),
      CodeBlockLowlight.configure({ lowlight }),
    ],
    content: '',
    editorProps: {
      attributes: { class: 'prose max-w-none min-h-[300px] outline-none' },
    },
  })

  const { data: existingBlog, isLoading } = useQuery({
    queryKey: ['blog', id],
    queryFn: () => blogsApi.getBlog(id!).then((r) => r.data),
    enabled: !!id,
  })

  useEffect(() => {
    if (!existingBlog) return
    setTitle(existingBlog.title)
    setPrivacy(existingBlog.privacy)
    setTags(existingBlog.tags.map((t) => t.name).join(', '))
    if (existingBlog.thumbnail_url) setThumbnail(existingBlog.thumbnail_url)
    editor?.commands.setContent(existingBlog.content)
  }, [existingBlog, editor])

  const handleImageUpload = useCallback(async (e: { target: { files: FileList | null } }) => {
    const file = e.target.files?.[0]
    if (!file) return
    try {
      const res = await blogsApi.uploadImage(file)
      editor?.chain().focus().setImage({ src: res.data.url }).run()
    } catch {
      setError('Image upload failed')
    }
  }, [editor])

  const save = async (status: 'draft' | 'published') => {
    if (!editor || !title.trim()) { setError('Title is required'); return }
    setSaving(true)
    setError('')
    const input: CreateBlogInput = {
      title,
      content: editor.getHTML(),
      privacy,
      status,
      tag_names: tags.split(',').map((t) => t.trim()).filter(Boolean),
      category_ids: [],
      ...(thumbnail ? { thumbnail_url: thumbnail } : {}),
    }
    try {
      const result = id
        ? await blogsApi.updateBlog(id, input)
        : await blogsApi.createBlog(input)
      navigate(`/blog/${result.data.id}`)
    } catch {
      setError('Failed to save. Please try again.')
    } finally {
      setSaving(false)
    }
  }

  if (id && isLoading) return <div className="animate-pulse h-96 bg-gray-100 rounded-lg" />

  return (
    <div className="max-w-3xl mx-auto">
      <h1 className="text-2xl font-bold mb-6">{id ? 'Edit post' : 'New post'}</h1>

      {error && (
        <div className="bg-red-50 text-red-600 text-sm px-3 py-2 rounded mb-4" role="alert">{error}</div>
      )}

      <div className="space-y-4">
        <input
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          placeholder="Post title…"
          className="w-full text-2xl font-semibold border-0 border-b border-gray-200 pb-2 focus:outline-none focus:border-blue-500"
          data-testid="title-input"
        />

        <div className="border border-gray-200 rounded-lg p-4">
          <div className="flex gap-2 mb-2 border-b pb-2">
            <button onClick={() => editor?.chain().focus().toggleBold().run()} className="text-sm font-bold px-2 py-1 rounded hover:bg-gray-100">B</button>
            <button onClick={() => editor?.chain().focus().toggleItalic().run()} className="text-sm italic px-2 py-1 rounded hover:bg-gray-100">I</button>
            <button onClick={() => editor?.chain().focus().toggleCode().run()} className="text-sm font-mono px-2 py-1 rounded hover:bg-gray-100">`</button>
            <button onClick={() => editor?.chain().focus().toggleCodeBlock().run()} className="text-sm font-mono px-2 py-1 rounded hover:bg-gray-100">```</button>
            <label className="text-sm px-2 py-1 rounded hover:bg-gray-100 cursor-pointer">
              📷
              <input type="file" accept="image/*" className="hidden" onChange={handleImageUpload} />
            </label>
          </div>
          <EditorContent editor={editor} data-testid="editor-content" />
        </div>

        <div className="flex gap-4">
          <div className="flex-1">
            <label className="block text-sm font-medium text-gray-700 mb-1">Tags (comma-separated)</label>
            <input
              value={tags}
              onChange={(e) => setTags(e.target.value)}
              placeholder="react, typescript, webdev"
              className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Privacy</label>
            <select
              value={privacy}
              onChange={(e) => setPrivacy(e.target.value as CreateBlogInput['privacy'])}
              className="border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              data-testid="privacy-select"
            >
              {PRIVACY_OPTIONS.map((o) => (
                <option key={o.value} value={o.value}>{o.label}</option>
              ))}
            </select>
          </div>
        </div>

        <div className="flex gap-3 justify-end">
          <button
            onClick={() => save('draft')}
            disabled={saving}
            className="border border-gray-300 px-4 py-2 rounded text-sm hover:bg-gray-50 disabled:opacity-60"
          >
            Save draft
          </button>
          <button
            onClick={() => save('published')}
            disabled={saving}
            className="bg-blue-600 text-white px-4 py-2 rounded text-sm hover:bg-blue-700 disabled:opacity-60"
            data-testid="publish-btn"
          >
            {saving ? 'Publishing…' : 'Publish'}
          </button>
        </div>
      </div>
    </div>
  )
}
