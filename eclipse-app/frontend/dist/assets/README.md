# 🎨 Eclipse Collection: Graphic Design & Asset Guide

Welcome! The launcher is built so you can drop your custom `.png` and `.mp3` files directly into these folders without writing a single line of code. The UI will instantly pick them up.

---

## 📁 Folder Structure

```text
eclipse-app/frontend/dist/assets/
├── audio/
│   ├── menu_theme.mp3        <-- Your Lament of Innocence theme music!
│   ├── cursor.mp3            <-- (Optional) Cursor click SFX
│   ├── confirm.mp3           <-- (Optional) Selection chime SFX
│   └── launch.mp3            <-- (Optional) Game start strike SFX
└── games/
    ├── sotn/                 <-- Symphony of the Night
    ├── rondo/                <-- Rondo of Blood
    ├── loi/                  <-- Lament of Innocence
    ├── cod/                  <-- Curse of Darkness
    ├── lod/                  <-- Legacy of Darkness
    ├── cv64/                 <-- Castlevania 64
    └── rebirth/              <-- The Adventure ReBirth (M2)
```

---

## 🖼️ Game Artwork Specifications

Inside each game folder (`games/sotn/`, `games/rondo/`, etc.), you can drop:

| File Name | Format | Recommended Dimensions | Description |
| :--- | :--- | :--- | :--- |
| **`box_us.png`** | Transparent / Flat PNG | ~600 × 840 px (Portrait) | North American cover / box art. Shown when "USA" region is selected. |
| **`box_jp.png`** | Transparent / Flat PNG | ~600 × 840 px (Portrait) | Japanese cover / box art. Shown when "JPN" region is selected. |
| **`box_eu.png`** | Transparent / Flat PNG | ~600 × 840 px (Portrait) | European cover / box art. Shown when "EUR" region is selected. |
| **`background.png`** | PNG or JPG | 1920 × 1080 px or 4K | Fullscreen cinematic background / stage wallpaper for this game. |
| **`logo.png`** | Transparent PNG | ~800 × 400 px | High-res transparent game title logo. |

---

## 🎵 Soundtrack & Audio Drop-in

Place your chosen track from the **Lament of Innocence** soundtrack in:
`eclipse-app/frontend/dist/assets/audio/menu_theme.mp3`

> [!NOTE]
> If an audio file or custom image is missing, the launcher will automatically fall back to its built-in gothic gold crest cards and real-time audio synthesizer, so nothing will ever crash or look broken!
