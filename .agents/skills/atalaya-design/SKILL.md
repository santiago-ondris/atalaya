---
name: atalaya-design
description: "This skill should be used when the user explicitly says 'Atalaya style', 'Atalaya design', '/atalaya-design', or directly asks to use/apply the Atalaya design system. NEVER trigger automatically for generic UI or design tasks."
version: 1.0.0
allowed-tools: [Read, Write, Edit, Glob, Grep]
---

# Atalaya

You are a senior product designer. When this skill is active, every UI decision follows this design language.

**Before starting any design work, declare the required fonts:** Instrument Serif 400, Public Sans 400/500/600, and IBM Plex Mono 400/600. Load them as documented in `references/platform-mapping.md`.

## 1. DESIGN PHILOSOPHY

Atalaya is a contemporary production watch station expressed through the language of a naval logbook. Operational precision sits on warm paper, brass rules, cartographic grids, plotted bearings, and geometric signal flags. The historical reference creates recognition and craft, but the interface must always remain faster to scan than the metaphor is to notice.

The defining tension is **modern operational instrumentation versus historical nautical materiality**. It is not a costume and never becomes skeuomorphic.

### Principles

1. **Status is a shape before it is a color.** Every critical state combines a signal flag, label, and color so it survives low vision and monochrome output.
2. **Paper carries work; green carries command.** Dense data lives on warm neutral surfaces while deep green anchors navigation, identity, and healthy state.
3. **Indigo means interaction.** Selection, focus, active flow, and primary actions use indigo; never use it as decoration.
4. **Rules replace shadows.** Depth comes from brass or neutral borders and adjacent surface changes; cards do not float.
5. **Serif announces, sans operates, mono records.** Instrument Serif is restricted to first impressions and major titles; operational UI uses Public Sans; machine facts use IBM Plex Mono.
6. **Density is organized, not reduced.** Show enough context to operate production, using consistent columns, 8px rhythm, and strong information hierarchy.
7. **The nautical reference stays geometric.** Use bearings, coordinates, ruled grids, flags, and concise logbook language; never decorative anchors, ship wheels, ropes, or waves.

## 2. CRAFT RULES — HOW TO COMPOSE

| Layer | Role | Treatment |
|---|---|---|
| 1 | Global command | Deep-green navigation and global health strip |
| 2 | Screen orientation | Instrument Serif title, concise timestamp, operational summary |
| 3 | Working surfaces | Warm paper, 1px rules, 6px maximum radius |
| 4 | Data and actions | Public Sans labels, IBM Plex Mono values, indigo selection |
| 5 | Urgent signal | Flag shape plus semantic color and explicit text |

- Use no more than two font families in one compact component; mono replaces body only for machine data.
- Compose on an 8px grid. Use 16px inside dense controls, 24px inside panels, and 48–64px only between major screen regions.
- Keep chromatic arrivals scarce: green for command/healthy, indigo for interaction, yellow for caution, brick red for failure, brass for structure.
- Prefer aligned rows and ruled sections to detached card collections. A dashboard should feel assembled, not sprinkled.
- Motion is mechanical: 120–180ms, ease-out character, no overshoot, no ambient movement in operational screens.
- Run the squint test: navigation, current screen, selected record, and critical state must remain distinguishable without reading copy.

## 3. ANTI-PATTERNS — WHAT TO NEVER DO

- No generic rounded SaaS cards; panel radius never exceeds 6px.
- No gradients in operational surfaces or buttons.
- No drop shadows on ordinary cards, tables, filters, or navigation.
- No glassmorphism, blur panels, neon glows, or translucent chrome.
- No decorative nautical clip art: anchors, compasses, wheels, ropes, waves, or boats.
- No Instrument Serif in tables, forms, buttons, filters, or body paragraphs.
- No green as a generic active color; active and selected controls are indigo.
- No status communicated by color alone; include flag geometry and text.
- No floating toast for durable operational failures; preserve them in a banner or event log.
- No skeleton shimmer. Loading uses a static ruled placeholder or compact spinner with a textual state.
- No hidden critical metadata behind hover-only tooltips.
- No empty card grid introduced only to fill space.

## 4. WORKFLOW

1. Declare and load fonts from `references/platform-mapping.md`.
2. Apply the semantic variables from `references/tokens.md`.
3. Build with the specifications in `references/components.md`.
4. Confirm status remains understandable without color.
5. Run the squint test and verify light and dark modes.
6. Test long labels, empty states, loading, failure, one row, and 100 rows.
7. Keep new values in `design-model.yaml` before using them in code.

## 5. REFERENCE FILES

| File | Contains |
|---|---|
| `references/tokens.md` | Typography, color, spacing, radii, motion, iconography |
| `references/components.md` | Operational components and their exact treatments |
| `references/platform-mapping.md` | Copy-ready CSS, React/Tailwind, and SwiftUI mappings |

