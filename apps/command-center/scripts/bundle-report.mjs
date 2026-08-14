import { readFileSync } from 'node:fs'
import { gzipSync } from 'node:zlib'
import { join } from 'node:path'

const output = join(process.cwd(), 'dist')
const manifest = JSON.parse(readFileSync(join(output, '.vite/manifest.json'), 'utf8'))
const entries = Object.entries(manifest)
const initial = entries.find(([, value]) => value.isEntry)
const scene = entries.find(([key]) => key.endsWith('/LighthouseScene.tsx'))

if (!initial || !scene)
  throw new Error('No se encontraron los bundles esperados en el manifest.')

function stats(files) {
  return files.map((file) => {
    const bytes = readFileSync(join(output, file))
    return { file, raw: bytes.length, gzip: gzipSync(bytes).length }
  })
}

function format(bytes) {
  return `${(bytes / 1024).toFixed(1)} kB`
}

for (const [label, entry] of [
  ['Inicial', initial[1]],
  ['LighthouseScene', scene[1]],
]) {
  const rows = stats([entry.file, ...(entry.css ?? [])])
  const raw = rows.reduce((sum, row) => sum + row.raw, 0)
  const gzip = rows.reduce((sum, row) => sum + row.gzip, 0)
  console.log(`${label}: ${format(raw)} bruto / ${format(gzip)} gzip`)
  for (const row of rows)
    console.log(`  ${row.file}: ${format(row.raw)} / ${format(row.gzip)} gzip`)
  if (label === 'LighthouseScene' && gzip > 3 * 1024 * 1024) {
    throw new Error('Los recursos transferidos de la escena superan el límite de 3 MB.')
  }
}
