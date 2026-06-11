import '@testing-library/jest-dom'
import { vi } from 'vitest'

// Mock axios globally
vi.mock('axios', () => ({
  default: {
    create: vi.fn(() => ({
      get: vi.fn(),
      post: vi.fn(),
      patch: vi.fn(),
      delete: vi.fn(),
      put: vi.fn(),
      interceptors: {
        request: { use: vi.fn() },
        response: { use: vi.fn() },
      },
    })),
  },
}))

// Global react-i18next mock — returns English text so existing text-based tests pass.
// Individual test files can override via vi.mocked(useTranslation).mockReturnValue(...)
const EN: Record<string, string> = {
  'nav.articles': 'Articles',
  'nav.portfolio': 'Portfolio',
  'nav.author': 'Author',
  'nav.categories': 'Categories',
  'nav.forums': 'Forums',
  'nav.write': 'Write',
  'nav.signIn': 'Sign in',
  'nav.getStarted': 'Get started',
  'nav.logout': 'Logout',
  'nav.admin': 'Admin',
  'blog.minRead': 'min read',
  'blog.signInToReadMore': 'Sign in to read more →',
  'blog.postNotFound': 'Post not found.',
  'blog.copyLink': 'Copy link',
  'blog.delete': 'Delete',
  'blog.deleteConfirm': 'Delete this post?',
  'blog.edit': 'Edit',
  'blog.commentsLabel': 'Comments',
  'blog.commentPlaceholder': 'Add a comment…',
  'blog.post': 'Post',
  'blog.translationUnavailable': 'Translation unavailable',
  'blog.translationNotice': 'Showing original Vietnamese content.',
}

vi.mock('react-i18next', () => ({
  useTranslation: vi.fn(() => ({
    t: (key: string) => EN[key] ?? key,
    i18n: { changeLanguage: vi.fn(), language: 'vi' },
  })),
  initReactI18next: { type: '3rdParty', init: vi.fn() },
  Trans: ({ children }: { children: unknown }) => children,
}))
