import { Command } from 'cmdk'
import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useRef,
  useState,
  type ReactNode,
} from 'react'
import { useNavigate } from 'react-router'
import { lighthouseDestinations } from '../../features/lighthouse/destinations'

interface CommandPaletteContextValue {
  openPalette: (opener?: HTMLElement) => void
}

const CommandPaletteContext = createContext<CommandPaletteContextValue | null>(null)

export function CommandPaletteProvider({ children }: { children: ReactNode }) {
  const navigate = useNavigate()
  const [open, setOpen] = useState(false)
  const returnFocusRef = useRef<HTMLElement | null>(null)

  const openPalette = useCallback((opener?: HTMLElement) => {
    returnFocusRef.current =
      opener ??
      (document.activeElement instanceof HTMLElement ? document.activeElement : null)
    setOpen(true)
  }, [])

  const changeOpen = useCallback((nextOpen: boolean) => {
    setOpen(nextOpen)
    if (!nextOpen) {
      window.setTimeout(() => {
        window.setTimeout(() => returnFocusRef.current?.focus(), 0)
      }, 0)
    }
  }, [])

  useEffect(() => {
    function onKeyDown(event: KeyboardEvent) {
      if (event.key.toLowerCase() !== 'k' || (!event.metaKey && !event.ctrlKey)) return
      event.preventDefault()
      if (open) changeOpen(false)
      else openPalette()
    }
    document.addEventListener('keydown', onKeyDown)
    return () => document.removeEventListener('keydown', onKeyDown)
  }, [changeOpen, open, openPalette])

  function selectDestination(route: string) {
    changeOpen(false)
    navigate(route)
  }

  return (
    <CommandPaletteContext.Provider value={{ openPalette }}>
      {children}
      <Command.Dialog
        open={open}
        onOpenChange={changeOpen}
        label="Buscar destino"
        className="command-palette"
        filter={(value, search, keywords = []) =>
          normalizeSearch([value, ...keywords].join(' ')).includes(
            normalizeSearch(search),
          )
            ? 1
            : 0
        }
      >
        <div className="command-palette-heading">
          <span>ATALAYA / NAVEGACIÓN</span>
          <kbd>Esc</kbd>
        </div>
        <Command.Input
          autoFocus
          aria-label="Buscar destino"
          placeholder="Buscar destino…"
        />
        <Command.List aria-label="Destinos disponibles">
          <Command.Empty>Sin resultados</Command.Empty>
          {(['Aplicaciones', 'Operación'] as const).map((group) => (
            <Command.Group key={group} heading={group}>
              {lighthouseDestinations
                .filter((destination) => destination.group === group)
                .map((destination) => (
                  <Command.Item
                    key={destination.id}
                    value={destination.label}
                    keywords={[destination.context, ...destination.keywords]}
                    onSelect={() => selectDestination(destination.route)}
                  >
                    <span>{destination.label}</span>
                    <small>{destination.context}</small>
                  </Command.Item>
                ))}
            </Command.Group>
          ))}
        </Command.List>
      </Command.Dialog>
    </CommandPaletteContext.Provider>
  )
}

export function CommandPaletteTrigger({ className }: { className?: string }) {
  const { openPalette } = useCommandPalette()
  return (
    <button
      type="button"
      className={['command-palette-trigger', className].filter(Boolean).join(' ')}
      onClick={(event) => openPalette(event.currentTarget)}
      aria-keyshortcuts="Meta+K Control+K"
    >
      <span>Ir a…</span>
      <kbd aria-hidden="true">⌘/Ctrl K</kbd>
    </button>
  )
}

function normalizeSearch(value: string) {
  return value
    .normalize('NFD')
    .replace(/[\u0300-\u036f]/g, '')
    .toLocaleLowerCase('es')
    .trim()
}

function useCommandPalette() {
  const context = useContext(CommandPaletteContext)
  if (!context) throw new Error('CommandPaletteTrigger requires CommandPaletteProvider')
  return context
}
