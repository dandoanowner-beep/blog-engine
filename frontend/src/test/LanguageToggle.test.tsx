import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import LanguageToggle from '../components/LanguageToggle'

const mockChangeLanguage = vi.fn()

vi.mock('react-i18next', () => ({
  useTranslation: vi.fn(() => ({
    t: (key: string) => key,
    i18n: { changeLanguage: mockChangeLanguage, language: 'vi' },
  })),
  initReactI18next: { type: '3rdParty', init: vi.fn() },
}))

import { useTranslation } from 'react-i18next'

describe('LanguageToggle', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(useTranslation).mockReturnValue({
      t: (key: string) => key,
      i18n: { changeLanguage: mockChangeLanguage, language: 'vi' },
    } as never)
  })

  it('shows VN when language is vi', () => {
    render(<LanguageToggle />)
    expect(screen.getByTestId('lang-toggle-btn')).toHaveTextContent('VN')
  })

  it('shows EN when language is en', () => {
    vi.mocked(useTranslation).mockReturnValue({
      t: (key: string) => key,
      i18n: { changeLanguage: mockChangeLanguage, language: 'en' },
    } as never)
    render(<LanguageToggle />)
    expect(screen.getByTestId('lang-toggle-btn')).toHaveTextContent('EN')
  })

  it('dropdown is hidden by default', () => {
    render(<LanguageToggle />)
    expect(screen.queryByTestId('lang-dropdown')).not.toBeInTheDocument()
  })

  it('opens dropdown when toggle button is clicked', async () => {
    render(<LanguageToggle />)
    await userEvent.click(screen.getByTestId('lang-toggle-btn'))
    expect(screen.getByTestId('lang-dropdown')).toBeInTheDocument()
  })

  it('dropdown shows both language options', async () => {
    render(<LanguageToggle />)
    await userEvent.click(screen.getByTestId('lang-toggle-btn'))
    expect(screen.getByText('English')).toBeInTheDocument()
    expect(screen.getByText('Tiếng Việt')).toBeInTheDocument()
  })

  it('calls changeLanguage("en") and closes dropdown when English is selected', async () => {
    render(<LanguageToggle />)
    await userEvent.click(screen.getByTestId('lang-toggle-btn'))
    await userEvent.click(screen.getByText('English'))
    expect(mockChangeLanguage).toHaveBeenCalledWith('en')
    expect(screen.queryByTestId('lang-dropdown')).not.toBeInTheDocument()
  })

  it('calls changeLanguage("vi") and closes dropdown when Tiếng Việt is selected', async () => {
    vi.mocked(useTranslation).mockReturnValue({
      t: (key: string) => key,
      i18n: { changeLanguage: mockChangeLanguage, language: 'en' },
    } as never)
    render(<LanguageToggle />)
    await userEvent.click(screen.getByTestId('lang-toggle-btn'))
    await userEvent.click(screen.getByText('Tiếng Việt'))
    expect(mockChangeLanguage).toHaveBeenCalledWith('vi')
    expect(screen.queryByTestId('lang-dropdown')).not.toBeInTheDocument()
  })
})
