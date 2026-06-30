import { useEffect, useRef } from 'react'

const TESSERACT_4D: [number, number, number, number][] = []
for (let bits = 0; bits < 16; bits++) {
  TESSERACT_4D.push([
    (bits & 1) ? 1 : -1,
    (bits & 2) ? 1 : -1,
    (bits & 4) ? 1 : -1,
    (bits & 8) ? 1 : -1,
  ])
}

const TESSERACT_EDGES: [number, number][] = []
for (let i = 0; i < 16; i++) {
  for (let j = i + 1; j < 16; j++) {
    let diff = 0
    for (let k = 0; k < 4; k++) {
      if (TESSERACT_4D[i][k] !== TESSERACT_4D[j][k]) diff++
    }
    if (diff === 1) TESSERACT_EDGES.push([i, j])
  }
}

const CANVAS_W = 960
const CANVAS_H = 720
const ROT_SPEED_YW = 0.012
const ROT_SPEED_XZ = 0.015
const ROT_SPEED_XY = 0.009
const ROT_SPEED_ZW = 0.011

function getThemeColors() {
  const styles = getComputedStyle(document.documentElement)
  return {
    bg: readCSSColor(styles, '--punkdom-bg', '#1a1a1a'),
    text: readCSSColor(styles, '--punkdom-text-muted', '#a3a3a3'),
    accent: readCSSColor(styles, '--punkdom-accent', '#a8adb7'),
  }
}

function readCSSColor(styles: CSSStyleDeclaration, name: string, fallback: string) {
  return styles.getPropertyValue(name).trim() || fallback
}

function rotate4D(
  v: [number, number, number, number],
  ayw: number, axz: number, axy: number, azw: number,
): [number, number, number, number] {
  let [x, y, z, w] = v
  // YW
  const c1 = Math.cos(ayw), s1 = Math.sin(ayw)
  ;[y, w] = [y * c1 - w * s1, y * s1 + w * c1]
  // XZ
  const c2 = Math.cos(axz), s2 = Math.sin(axz)
  ;[x, z] = [x * c2 - z * s2, x * s2 + z * c2]
  // XY
  const c3 = Math.cos(axy), s3 = Math.sin(axy)
  ;[x, y] = [x * c3 - y * s3, x * s3 + y * c3]
  // ZW
  const c4 = Math.cos(azw), s4 = Math.sin(azw)
  ;[z, w] = [z * c4 - w * s4, z * s4 + w * c4]
  return [x, y, z, w]
}

function project4Dto3D(v: [number, number, number, number], d: number): [number, number, number] {
  const [x, y, z, w] = v
  const s = d / (d - w)
  return [x * s, y * s, z * s]
}

function project3Dto2D(v: [number, number, number], scale: number, cx: number, cy: number): [number, number, number] {
  return [v[0] * scale + cx, -v[1] * scale + cy, v[2] * scale]
}

function drawAsciiLine(
  grid: string[][], zbuf: number[][],
  x1: number, y1: number, z1: number,
  x2: number, y2: number, z2: number,
  cw: number, ch: number, cols: number, rows: number,
) {
  const dx = Math.abs(x2 - x1), dy = Math.abs(y2 - y1)
  const steps = Math.max(Math.round(dx / cw * 1.5), Math.round(dy / ch * 1.5), 4)
  const segDx = x2 - x1, segDy = y2 - y1
  const absDx = Math.abs(segDx), absDy = Math.abs(segDy)

  let lineCh: string
  if (absDx > absDy * 2) lineCh = '-'
  else if (absDy > absDx * 2) lineCh = '|'
  else if (segDx * segDy > 0) lineCh = '\\'
  else lineCh = '/'

  for (let i = 0; i <= steps; i++) {
    const t = i / steps
    const x = x1 + (x2 - x1) * t
    const y = y1 + (y2 - y1) * t
    const z = z1 + (z2 - z1) * t
    const col = Math.floor(x / cw)
    const row = Math.floor(y / ch)
    if (col < 0 || col >= cols || row < 0 || row >= rows) continue
    if (z > zbuf[row][col]) {
      zbuf[row][col] = z
      grid[row][col] = lineCh
    }
  }
}

function colorToRGB(color: string, fallback: [number, number, number]): [number, number, number] {
  const hex = color.match(/^#([0-9a-f]{3}|[0-9a-f]{6})$/i)
  if (hex) {
    const value = hex[1]
    if (value.length === 3) {
      return [
        parseInt(value[0] + value[0], 16),
        parseInt(value[1] + value[1], 16),
        parseInt(value[2] + value[2], 16),
      ]
    }
    return [
      parseInt(value.slice(0, 2), 16),
      parseInt(value.slice(2, 4), 16),
      parseInt(value.slice(4, 6), 16),
    ]
  }

  const rgb = color.match(/^rgba?\(\s*(\d{1,3})[\s,]+(\d{1,3})[\s,]+(\d{1,3})/i)
  if (rgb) {
    return [
      clampColorChannel(Number(rgb[1])),
      clampColorChannel(Number(rgb[2])),
      clampColorChannel(Number(rgb[3])),
    ]
  }

  return fallback
}

function clampColorChannel(value: number) {
  if (!Number.isFinite(value)) return 0
  return Math.max(0, Math.min(255, Math.round(value)))
}

export function HyperCube() {
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const aywRef = useRef(0.3)
  const axzRef = useRef(0.5)
  const axyRef = useRef(0.7)
  const azwRef = useRef(0.4)
  const rafRef = useRef(0)

  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) return
    let ctx: CanvasRenderingContext2D | null = null
    try {
      ctx = canvas.getContext('2d')
    } catch {
      ctx = null
    }
    if (!ctx) return

    const gridCols = 80
    const gridRows = Math.floor(gridCols * CANVAS_H / CANVAS_W)
    const cw = CANVAS_W / gridCols
    const ch = CANVAS_H / gridRows
    const cx = CANVAS_W / 2
    const cy = CANVAS_H * 0.48
    const scale = 165
    const perspDist = 4.5

    let lastTime = performance.now()
    const charGrid = Array.from({ length: gridRows }, () => new Array(gridCols).fill(' '))
    const depthGrid = Array.from({ length: gridRows }, () => new Array(gridCols).fill(-99999))

    const draw = () => {
      const now = performance.now()
      const dt = Math.min(now - lastTime || 16, 100)
      lastTime = now

      const fpsFactor = Math.min(dt / (1000 / 24), 3)
      aywRef.current += ROT_SPEED_YW * fpsFactor
      axzRef.current += ROT_SPEED_XZ * fpsFactor
      axyRef.current += ROT_SPEED_XY * fpsFactor
      azwRef.current += ROT_SPEED_ZW * fpsFactor

      const colors = getThemeColors()

      // 清空 buffer
      for (let r = 0; r < gridRows; r++) {
        charGrid[r].fill(' ')
        depthGrid[r].fill(-99999)
      }

      // 4D → 3D → 2D
      const proj: [number, number, number][] = TESSERACT_4D.map(v => {
        const r4 = rotate4D(v, aywRef.current, axzRef.current, axyRef.current, azwRef.current)
        const p3 = project4Dto3D(r4, perspDist)
        return project3Dto2D(p3, scale, cx, cy)
      })

      // 画线段
      for (const [a, b] of TESSERACT_EDGES) {
        drawAsciiLine(charGrid, depthGrid, ...proj[a], ...proj[b], cw, ch, gridCols, gridRows)
      }

      // 画顶点
      for (let vi = 0; vi < proj.length; vi++) {
        const [px, py, pz] = proj[vi]
        const col = Math.floor(px / cw)
        const row = Math.floor(py / ch)
        if (col >= 0 && col < gridCols && row >= 0 && row < gridRows) {
          if (pz > depthGrid[row][col]) {
            depthGrid[row][col] = pz
            charGrid[row][col] = 'o'
          }
        }
      }

      // 渲染到 Canvas
      ctx.fillStyle = colors.bg
      ctx.fillRect(0, 0, CANVAS_W, CANVAS_H)

      ctx.font = `bold ${ch * 0.92}px "Courier New", monospace`
      ctx.textAlign = 'center'
      ctx.textBaseline = 'middle'

      // 计算深度范围
      const depths: number[] = []
      for (let r = 0; r < gridRows; r++)
        for (let c = 0; c < gridCols; c++)
          if (depthGrid[r][c] > -99999) depths.push(depthGrid[r][c])
      const dMax = depths.length > 0 ? Math.max(...depths) : 10
      const dMin = depths.length > 0 ? Math.min(...depths) : -10

      const fg = colorToRGB(colors.text, [83, 96, 113])
      const accent = colorToRGB(colors.accent, [168, 173, 183])

      for (let row = 0; row < gridRows; row++) {
        for (let col = 0; col < gridCols; col++) {
          const glyph = charGrid[row][col]
          if (glyph === ' ') continue
          const x = col * cw + cw / 2
          const y = row * ch + ch / 2
          const depth = depthGrid[row][col]
          const nd = dMax - dMin > 0 ? (depth - dMin) / (dMax - dMin) : 0.5
          const alpha = 0.25 + nd * 0.75
          const [cr, cg, cb] = glyph === 'o' ? accent : fg
          ctx.fillStyle = `rgba(${cr},${cg},${cb},${alpha})`
          ctx.fillText(glyph, x, y)
        }
      }

      rafRef.current = requestAnimationFrame(draw)
    }

    rafRef.current = requestAnimationFrame(draw)
    return () => cancelAnimationFrame(rafRef.current)
  }, [])

  return (
    <canvas
      ref={canvasRef}
      width={CANVAS_W}
      height={CANVAS_H}
      className="block bg-[var(--punkdom-bg)]"
      style={{
        width: 'auto',
        height: 'auto',
        maxWidth: '100%',
        maxHeight: '100%',
        aspectRatio: `${CANVAS_W} / ${CANVAS_H}`,
      }}
      aria-hidden="true"
    />
  )
}
