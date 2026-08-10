# Atalaya — Components

Every component below is derived from the approved design model. “Observed” means it already existed in the foundation screen; “derived” means it follows the confirmed principles.

## 1. BUTTONS (derived)

| Variant | Background | Text | Border | Radius | Height |
|---|---|---|---|---:|---:|
| Primary | `--accent` | white | 1px solid `--accent` | 4px | 40px |
| Secondary | transparent | `--text1` | 1px solid `--border-visible` | 4px | 40px |
| Ghost | transparent | `--text2` | none | 4px | 36px |
| Destructive | `--error` | white | 1px solid `--error` | 4px | 40px |

Padding is 10px 16px, Public Sans 13px/600. Hover darkens or adds `--accent-subtle`; pressed translates 1px down; disabled uses `--text4` and no pointer events. Focus is a 2px `--accent` outline with 2px offset.

## 2. CARDS / SURFACES (observed, refined)

Standard panels use `--surface1`, 1px `--border`, 6px radius, 24px padding, no shadow. Featured panels replace one edge with a 2px `--border-visible` rule. Compact panels use 16px padding. Titles use `--text-subheading`; descriptions use body-sm and `--text2`; metadata uses mono caption.

## 3. INPUTS (derived)

Height 40px, `--surface1`, 1px `--border-visible`, 4px radius, 10px 12px padding. Focus changes the border and a 2px outline to `--accent`. Error changes border to `--error` and adds persistent caption text. Labels sit 8px above in mono uppercase `--text-label`. Search inputs may contain one 18px Tabler icon.

## 4. LISTS / DATA ROWS (derived)

Rows are at least 48px high with 8px 16px padding and a bottom `--border` rule. Labels use body-sm; values, timestamps, identifiers, and counts use IBM Plex Mono. Hover uses `--surface2`; selected uses `--accent-subtle` plus a 2px indigo leading rule. Never remove the text label when a row is selected.

## 5. NAVIGATION (derived)

Desktop sidebar is 240px wide, deep green, white text, and no shadow. The wordmark uses Instrument Serif; navigation items use Public Sans 13px/500. Active navigation uses a parchment rectangle with black text and 4px radius. The global health strip is 40px high and uses compact mono facts. Below 760px, replace the sidebar with a top bar and compact bottom navigation.

## 6. TAGS AND FLAGS

Text tags are 24px tall, 2px radius, 4px 8px padding, mono caption, and a 1px current-color border. Status flags are custom 30×22px hard-edged geometric shapes with a black/white outline. Healthy is a green/paper diagonal, warning a yellow/black vertical split, critical a brick-red field with a paper square. Always place readable status text adjacent.

## 7. OVERLAYS

Dialogs use `--surface1`, a 1px brass border, 6px radius, 24px padding, and level-3 shadow. Backdrop is `rgba(13,16,14,.66)`. Maximum width is 560px. Sheets are reserved for mobile filters; desktop filtering stays inline.

## 8. TABLES AND PAGINATION

Table headers use mono uppercase labels on `--surface2`, 40px high. Rows are 48–56px. Numeric columns align right and use mono; message columns truncate only when a detail route is available. Pagination uses square 36px controls; current page is indigo with white text. Always show total records and current range.

## 9. FILTERS

Filters form one ruled toolbar, not independent pills. Selects and date controls are 40px high. Active filters receive `--accent-subtle`; a visible “Limpiar filtros” action appears when any filter differs from default. Filter state belongs in the URL.

## 10. STATE PATTERNS

- **Loading:** static ruled placeholders plus “Consultando bitácora…”; no shimmer.
- **Empty:** one signal flag, an exact explanation, and a relevant action if recovery is possible.
- **Error:** persistent error banner with source, timestamp, retry action, and correlation ID.
- **Disabled:** opacity is not the sole cue; use disabled color, cursor, and explanatory text when needed.
- **Stale:** amber flag and last-successful-update timestamp remain visible.

## 11. TOOLTIP AND ALERTS

Tooltips use dark `--text1` inverse surfaces, 2px radius, 8px padding, 240px maximum width, and appear after 350ms. They never contain critical-only information. Alerts use 6px radius, 16px padding, a 4px semantic leading rule, title/body/action layout, and remain until their condition clears or the user explicitly dismisses them.

## 12. PROGRESS AND LOADING INDICATORS

Progress bars are 6px high with 2px radius and rounded stroke caps. Track uses `--surface3`; fills use semantic colors. Spinner sizes are 16/24/32px, 2px stroke, `--accent`, and rotate linearly in 800ms. Skeletons are static `--surface2` blocks and never shimmer.

