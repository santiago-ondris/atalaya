# Atalaya — Platform Mapping

## 1. WEB FONT LOADING

```html
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=IBM+Plex+Mono:wght@400;600&family=Instrument+Serif:ital@0;1&family=Public+Sans:wght@400;500;600&display=swap" rel="stylesheet">
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@tabler/icons-webfont@3.41.1/dist/tabler-icons.min.css">
```

## 2. CSS CUSTOM PROPERTIES

```css
:root {
  --background:#F1F0E8; --bg:var(--background);
  --surface1:#FAFAE4; --surface2:#DEDCCF; --surface3:#B9B7AA;
  --border:#DEDCCF; --border-visible:#BA8A40;
  --text1:#000000; --text2:#51534E; --text3:#6B6C65; --text4:#8B8A80;
  --accent:#4750A8; --accent-subtle:#EEF2FF;
  --success:#005838; --success-bg:#E7F3ED;
  --warning:#FFE000; --warning-bg:#FFF9CC;
  --error:#C1432E; --error-bg:#FBE9E4;
  --font-display:"Instrument Serif",Georgia,serif;
  --font-body:"Public Sans",Arial,sans-serif;
  --font-mono:"IBM Plex Mono",Consolas,monospace;
  --text-display:64px; --text-heading:36px; --text-subheading:18px;
  --text-body:15px; --text-body-sm:13px; --text-caption:12px; --text-label:11px;
  --space-2xs:2px; --space-xs:4px; --space-sm:8px; --space-md:16px;
  --space-lg:24px; --space-xl:32px; --space-2xl:48px; --space-3xl:64px; --space-4xl:96px;
  --radius-element:2px; --radius-control:4px; --radius-component:6px; --radius-container:6px; --radius-pill:999px;
  --duration-fast:120ms; --duration-medium:160ms; --duration-slow:180ms;
  --ease-mechanical:cubic-bezier(.2,.8,.2,1);
  --shadow-1:none; --shadow-2:0 8px 0 rgba(0,0,0,.10); --shadow-3:0 16px 40px rgba(0,0,0,.22);
}
[data-theme="dark"] {
  --background:#0D100E; --bg:var(--background);
  --surface1:#171B18; --surface2:#242824; --surface3:#393C38;
  --border:#242824; --border-visible:#BA8A40;
  --text1:#FFFFFF; --text2:#B9B7AA; --text3:#8B8A80; --text4:#51534E;
  --accent:#979FDD; --accent-subtle:#11152E;
  --success:#005838; --success-bg:#003321;
  --warning:#FFE000; --warning-bg:#574C00;
  --error:#C1432E; --error-bg:#642217;
}
```

## 3. REACT / TAILWIND

Use CSS variables as the source of truth. A compatible Tailwind extension is:

```js
export default {
  content: ['./src/**/*.{js,ts,jsx,tsx}'],
  theme: { extend: {
    colors: {
      background:'var(--background)', surface:{1:'var(--surface1)',2:'var(--surface2)',3:'var(--surface3)'},
      border:{DEFAULT:'var(--border)',visible:'var(--border-visible)'},
      text:{1:'var(--text1)',2:'var(--text2)',3:'var(--text3)',4:'var(--text4)'},
      accent:{DEFAULT:'var(--accent)',subtle:'var(--accent-subtle)'},
      success:{DEFAULT:'var(--success)',bg:'var(--success-bg)'},
      warning:{DEFAULT:'var(--warning)',bg:'var(--warning-bg)'},
      error:{DEFAULT:'var(--error)',bg:'var(--error-bg)'}
    },
    fontFamily:{display:['Instrument Serif','Georgia','serif'],body:['Public Sans','Arial','sans-serif'],mono:['IBM Plex Mono','Consolas','monospace']},
    spacing:{'2xs':'2px',xs:'4px',sm:'8px',md:'16px',lg:'24px',xl:'32px','2xl':'48px','3xl':'64px','4xl':'96px'},
    borderRadius:{element:'2px',control:'4px',component:'6px',container:'6px',pill:'999px'},
    transitionDuration:{fast:'120ms',medium:'160ms',slow:'180ms'},
    transitionTimingFunction:{mechanical:'cubic-bezier(.2,.8,.2,1)'}
  }}, plugins: []
}
```

Prefer component CSS or CSS Modules for Atalaya's custom signal flags and cartographic stage; these should not be reconstructed from utility-class strings.

## 4. SWIFTUI

The current product is web-first. If a native companion is built, register Instrument Serif, Public Sans, and IBM Plex Mono as bundled fonts, mirror the semantic colors in an Asset Catalog with light/dark appearances, use continuous 2/4/6px radii, and preserve the 8px spacing grid. Do not substitute SF fonts silently because typography is brand-bearing.
