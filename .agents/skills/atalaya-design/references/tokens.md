# Atalaya — Tokens

## 0. PRIMITIVES

The neutral ramp is warm and slightly green-gray so paper and dark control-room modes belong to the same material family.

| Step | Neutral | Indigo brand |
|---|---|---|
| 50 | `#FAFAE4` | `#EEF2FF` |
| 100 | `#F1F0E8` | `#DDE2FF` |
| 200 | `#DEDCCF` | `#BEC6F4` |
| 300 | `#B9B7AA` | `#979FDD` |
| 400 | `#8B8A80` | `#7079C2` |
| 500 | `#6B6C65` | `#4750A8` |
| 600 | `#51534E` | `#3C4596` |
| 700 | `#393C38` | `#313979` |
| 800 | `#242824` | `#272E60` |
| 900 | `#171B18` | `#1D234A` |
| 950 | `#0D100E` | `#11152E` |

| Status | 50 | 500 | 900 |
|---|---|---|---|
| Green | `#E7F3ED` | `#005838` | `#003321` |
| Amber | `#FFF9CC` | `#FFE000` | `#574C00` |
| Red | `#FBE9E4` | `#C1432E` | `#642217` |
| Brass | `#F7F0E4` | `#BA8A40` | `#513915` |

Spacing primitives: `0, 2, 4, 8, 16, 24, 32, 48, 64, 96`.

Radii primitives: `0, 2, 4, 6, 999`.

## 1. TYPOGRAPHY

| Role | Font | Fallback | Weight | Use |
|---|---|---|---|---|
| Display | Instrument Serif | Georgia, serif | 400 | Login statement, screen title, app names |
| Body/UI | Public Sans | Arial, sans-serif | 400/500/600 | Navigation, copy, controls, tables |
| Mono/data | IBM Plex Mono | Consolas, monospace | 400/600 | IDs, timestamps, metrics, endpoints, traces |

`mono_for_code: true` and `mono_for_metrics: true`. Production facts should look recorded, not marketed. Instrument Serif is an approved observed brand decision, not an invented default.

| Token | Size | Line height | Tracking | Weight | Use |
|---|---:|---:|---:|---:|---|
| `--text-display` | 64px | .95 | -.035em | 400 | Entry statement |
| `--text-heading` | 36px | 1.05 | -.02em | 400 | Screen heading |
| `--text-subheading` | 18px | 1.25 | -.01em | 600 | Section/card title |
| `--text-body` | 15px | 1.5 | 0 | 400 | UI and prose |
| `--text-body-sm` | 13px | 1.45 | 0 | Dense rows |
| `--text-caption` | 12px | 1.35 | .01em | Timestamps and notes |
| `--text-label` | 11px | 1.2 | .12em | Uppercase metadata labels |

Use at most one serif title per viewport. Labels are uppercase mono. Never set paragraphs, buttons, or tables in Instrument Serif.

## 2. COLOR SYSTEM

| Token | Light | Dark | Role |
|---|---|---|---|
| `--background` / `--bg` | `#F1F0E8` | `#0D100E` | Canvas |
| `--surface1` | `#FAFAE4` | `#171B18` | Primary work surface |
| `--surface2` | `#DEDCCF` | `#242824` | Grouped surface |
| `--surface3` | `#B9B7AA` | `#393C38` | Inset/disabled well |
| `--border` | `#DEDCCF` | `#242824` | Quiet rule |
| `--border-visible` | `#BA8A40` | `#BA8A40` | Intentional rule |
| `--text1` | `#000000` | `#FFFFFF` | Primary text |
| `--text2` | `#51534E` | `#B9B7AA` | Secondary text |
| `--text3` | `#6B6C65` | `#8B8A80` | Metadata |
| `--text4` | `#8B8A80` | `#51534E` | Disabled |
| `--accent` | `#4750A8` | `#979FDD` | Selection/focus/action |
| `--accent-subtle` | `#EEF2FF` | `#11152E` | Selected background |
| `--success` | `#005838` | `#005838` | Healthy |
| `--warning` | `#FFE000` | `#FFE000` | Attention |
| `--error` | `#C1432E` | `#C1432E` | Critical |
| `--success-bg` | `#E7F3ED` | `#003321` | Healthy tint |
| `--warning-bg` | `#FFF9CC` | `#574C00` | Warning tint |
| `--error-bg` | `#FBE9E4` | `#642217` | Error tint |

Green is command and health, indigo is interaction, brass is structure. Never use amber text on paper; pair its yellow field with black text.

## 3. SPACING

| Token | Value | Use |
|---|---:|---|
| `--space-2xs` | 2px | Optical correction |
| `--space-xs` | 4px | Icon gaps |
| `--space-sm` | 8px | Tight internal gap |
| `--space-md` | 16px | Standard padding |
| `--space-lg` | 24px | Panel padding |
| `--space-xl` | 32px | Section cluster |
| `--space-2xl` | 48px | Major break |
| `--space-3xl` | 64px | Screen division |
| `--space-4xl` | 96px | Entry-stage breathing room |

## 4. BORDERS & RADII

| Token | Value | Use |
|---|---:|---|
| `--radius-element` | 2px | Flags, checks, tags |
| `--radius-control` | 4px | Buttons and inputs |
| `--radius-component` | 6px | Cards and panels |
| `--radius-container` | 6px | Dialogs and menus |
| `--radius-pill` | 999px | Status dot only; text chips stay square |

All standard surfaces use a 1px `--border` rule. Inputs and meaningful boundaries use `--border-visible`. Corners are machined, never soft or bubbly.

## 5. ELEVATION

The primary strategy is flat. Level 0 and ordinary cards use no shadow; menus may use `0 8px 0 rgba(0,0,0,.10)`; modals use `0 16px 40px rgba(0,0,0,.22)`. A border must remain visible even when a shadow exists.

## 6. MOTION

Mechanical, short, and causal. Micro interactions use `120ms`; content transitions `160ms`; dialogs `180ms`; easing is `cubic-bezier(.2,.8,.2,1)`. No bounce, parallax, ambient drift, or shimmer.

## 7. ICONOGRAPHY

Observed: 1.5px technical outlines with soft terminals and geometric construction, combined with solid custom signal flags. Use **Tabler Icons outline 1.5px** as the single functional kit. It matches technical density better than Lucide and provides the breadth required for observability.

CDN: `https://cdn.jsdelivr.net/npm/@tabler/icons-webfont@3.41.1/dist/tabler-icons.min.css`; usage: `ti ti-{name}`. Use 16px inline, 18px in buttons, and 20px in navigation. Icons inherit current text color. Tabler is a fallback; Atalaya's signal flags are custom components and must not be replaced with kit icons.

