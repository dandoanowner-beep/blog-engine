import { useState, useRef, useEffect } from 'react'
import { useTranslation } from 'react-i18next'

const langs = [
  { code: 'vi', label: 'Tiếng Việt', display: 'VN' },
  { code: 'en', label: 'English',    display: 'EN' },
] as const

function GlobeIcon() {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" className="w-4 h-4" viewBox="0 0 24 24" fill="none"
      stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <circle cx="12" cy="12" r="10" />
      <line x1="2" y1="12" x2="22" y2="12" />
      <path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z" />
    </svg>
  )
}

export default function LanguageToggle() {
  const { i18n } = useTranslation()
  const [isOpen, setIsOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setIsOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  const current = langs.find((l) => l.code === i18n.language) ?? langs[0]

  return (
    <div className="relative" ref={ref}>
      <button
        data-testid="lang-toggle-btn"
        onClick={() => setIsOpen((o) => !o)}
        className="flex items-center gap-1 text-sm text-gray-600 hover:text-blue-600 transition-colors"
      >
        <GlobeIcon />
        <span className="text-xs font-semibold">{current.display}</span>
        <svg xmlns="http://www.w3.org/2000/svg" className="w-3 h-3" viewBox="0 0 24 24" fill="none"
          stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
          <polyline points="6 9 12 15 18 9" />
        </svg>
      </button>

      {isOpen && (
        <div
          data-testid="lang-dropdown"
          className="absolute right-0 mt-1 min-w-[140px] bg-white border border-gray-200 rounded-lg shadow-lg z-50 overflow-hidden"
        >
          {langs.map(({ code, label, display }) => (
            <button
              key={code}
              onClick={() => { i18n.changeLanguage(code); setIsOpen(false) }}
              className={`w-full flex items-center gap-2 px-3 py-2 text-sm hover:bg-gray-50 transition-colors ${
                i18n.language === code ? 'text-blue-600 font-medium' : 'text-gray-700'
              }`}
            >
              <GlobeIcon />
              <span className="font-semibold">{display}</span>
              <span className="text-gray-500">{label}</span>
              {i18n.language === code && (
                <svg xmlns="http://www.w3.org/2000/svg" className="w-3.5 h-3.5 ml-auto text-blue-600"
                  viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5"
                  strokeLinecap="round" strokeLinejoin="round">
                  <polyline points="20 6 9 17 4 12" />
                </svg>
              )}
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
