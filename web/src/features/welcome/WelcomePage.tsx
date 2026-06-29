import { useCallback, useEffect } from 'react'
import { useWorkspaceStore, WELCOME_DISMISSED_KEY } from '@/stores/workspace-store'
import { HyperCube } from './HyperCube'
import { AsciiLogo } from './AsciiLogo'

export function WelcomePage() {
  const setShowWelcome = useWorkspaceStore((s) => s.setShowWelcome)

  const dismiss = useCallback(() => {
    if (typeof window !== 'undefined') {
      window.sessionStorage.setItem(WELCOME_DISMISSED_KEY, 'true')
    }
    setShowWelcome(false)
  }, [setShowWelcome])

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') dismiss()
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [dismiss])

  return (
    <button
      type="button"
      onClick={dismiss}
      className="fixed inset-0 z-[100] flex cursor-pointer select-none flex-col items-center justify-center border-0 bg-[var(--punkdom-bg)] p-0 text-[var(--punkdom-text)] outline-none"
      aria-label="Enter Punkdom"
    >
      <div className="flex h-full w-full -translate-y-[clamp(20px,4vh,44px)] flex-col items-center justify-center gap-[clamp(8px,1.5vh,18px)] px-4 py-8">
        <div className="flex w-full max-w-[min(86vw,86vh,860px)] items-center justify-center">
          <HyperCube />
        </div>
        <div className="-mt-[clamp(4px,1.5vh,16px)]">
          <AsciiLogo />
        </div>
      </div>
    </button>
  )
}
