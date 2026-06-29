import type { CSSProperties } from 'react'

const LOGO_ART = [
  '    ____              __       __              ',
  '   / __ \\__  ______  / /______/ /___  ____ ___ ',
  '  / /_/ / / / / __ \\/ //_/ __  / __ \\/ __ `__ \\',
  ' / ____/ /_/ / / / / ,< / /_/ / /_/ / / / / / /',
  '/_/    \\__,_/_/ /_/_/|_|\\__,_/\\____/_/ /_/ /_/ ',
].join('\n')

type LogoCharStyle = CSSProperties & {
  '--punkdom-logo-alpha-min': string
  '--punkdom-logo-alpha-max': string
}

export function AsciiLogo() {
  return (
    <pre
      className="punkdom-ascii-logo max-w-full overflow-hidden whitespace-pre text-center font-mono text-[clamp(6px,1.45vw,12px)] leading-[1.15] text-[var(--punkdom-text)]"
      aria-label="Punkdom"
    >
      {Array.from(LOGO_ART).map((char, index) => {
        if (char === '\n') return '\n'
        if (char === ' ') return ' '
        const jitter = stableUnit(index, char)
        const minAlpha = 0.34 + jitter * 0.14
        const maxAlpha = Math.min(minAlpha + 0.12 + stableUnit(index + 17, char) * 0.08, 0.62)
        const delay = -stableUnit(index + 31, char) * 3.8
        return (
          <span
            key={`${index}-${char}`}
            className="punkdom-ascii-logo-char"
            style={{
              '--punkdom-logo-alpha-min': minAlpha.toFixed(3),
              '--punkdom-logo-alpha-max': maxAlpha.toFixed(3),
              animationDelay: `${delay.toFixed(2)}s`,
            } as LogoCharStyle}
          >
            {char}
          </span>
        )
      })}
    </pre>
  )
}

function stableUnit(index: number, char: string) {
  const code = char.charCodeAt(0)
  const x = Math.sin((index + 1) * 12.9898 + code * 78.233) * 43758.5453
  return x - Math.floor(x)
}
