<template>
  <div class="color-picker-popup" ref="containerRef">
    <button class="color-trigger" @click="toggleOpen" :style="{ background: modelValue }">
      <span class="color-trigger-text">{{ modelValue }}</span>
    </button>

    <Teleport to="body">
      <div v-if="open" class="color-picker-dropdown" :style="dropdownStyle" ref="dropdownRef">
        <div class="color-picker-body">
          <div class="sv-palette" ref="svPaletteRef" @mousedown="startSVDrag">
            <div class="sv-palette-white" :style="{ background: `linear-gradient(90deg, #fff, hsl(${hue}, 100%, 50%))` }"></div>
            <div class="sv-palette-black" style="background: linear-gradient(rgba(0,0,0,0), #000)"></div>
            <div class="sv-handle" :style="{ left: svX + '%', top: svY + '%', background: modelValue }"></div>
          </div>

          <div class="sliders-wrap">
            <div class="hue-slider" ref="hueSliderRef" @mousedown="startHueDrag">
              <div class="hue-handle" :style="{ left: huePercent + '%' }"></div>
            </div>
            <div class="alpha-slider" ref="alphaSliderRef" @mousedown="startAlphaDrag">
              <div class="alpha-slider-bg" :style="{ background: `linear-gradient(90deg, transparent, ${hexColor})` }"></div>
              <div class="alpha-handle" :style="{ left: alphaPercent + '%' }"></div>
            </div>
          </div>

          <div class="preset-colors">
            <button
              v-for="c in presetColors"
              :key="c"
              class="preset-dot"
              :class="{ active: modelValue.toLowerCase() === c.toLowerCase() }"
              :style="{ background: c }"
              @click="selectColor(c)"
            ></button>
          </div>

          <div class="hex-input-row">
            <span class="hex-label">HEX</span>
            <input
              class="hex-input"
              :value="modelValue"
              @input="handleHexInput"
              @blur="handleHexBlur"
              maxlength="9"
            />
          </div>
        </div>
      </div>
    </Teleport>

    <div v-if="open" class="color-picker-backdrop" @click="open = false"></div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted, nextTick } from 'vue'

const props = defineProps<{
  modelValue: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const open = ref(false)
const containerRef = ref<HTMLElement | null>(null)
const dropdownRef = ref<HTMLElement | null>(null)
const svPaletteRef = ref<HTMLElement | null>(null)
const hueSliderRef = ref<HTMLElement | null>(null)
const alphaSliderRef = ref<HTMLElement | null>(null)

const hue = ref(210)
const svX = ref(60)
const svY = ref(40)
const alpha = ref(1)
const dropdownStyle = ref<Record<string, string>>({})

const presetColors = [
  '#2d8cf0', '#2d6cb4', '#667eea', '#f5576c', '#11998e',
  '#f7971e', '#ff9a9e', '#1a1a4e', '#fc00ff', '#e74c3c',
  '#3498db', '#2ecc71', '#f39c12', '#9b59b6', '#1abc9c',
  '#e67e22', '#34495e', '#7f8c8d', '#c0392b', '#16a085',
  '#27ae60', '#2980b9', '#8e44ad', '#d35400', '#2c3e50',
  '#f1c40f', '#00bcd4', '#ff5722', '#795548', '#607d8b',
]

const hexColor = computed(() => {
  const rgb = hslToRgb(hue.value, svX.value, 100 - svY.value)
  return rgbToHex(rgb.r, rgb.g, rgb.b)
})

const huePercent = computed(() => (hue.value / 360) * 100)
const alphaPercent = computed(() => alpha.value * 100)

function toggleOpen() {
  open.value = !open.value
  if (open.value) {
    parseColor(props.modelValue)
    nextTick(() => updateDropdownPosition())
  }
}

function updateDropdownPosition() {
  if (!containerRef.value) return
  const rect = containerRef.value.getBoundingClientRect()
  dropdownStyle.value = {
    position: 'fixed',
    top: rect.bottom + 8 + 'px',
    left: Math.min(rect.left, window.innerWidth - 280) + 'px',
    zIndex: '10002',
  }
}

function parseColor(hex: string) {
  const clean = hex.replace('#', '')
  if (clean.length === 8) {
    alpha.value = parseInt(clean.slice(6, 8), 16) / 255
  } else {
    alpha.value = 1
  }
  const rgb = hexToRgb(hex)
  if (rgb) {
    const hsl = rgbToHsl(rgb.r, rgb.g, rgb.b)
    hue.value = Math.round(hsl.h)
    svX.value = Math.round(hsl.s)
    svY.value = Math.round(100 - hsl.l)
  }
}

function selectColor(hex: string) {
  emitColor(hex)
  open.value = false
}

function emitColor(hex: string) {
  let final = hex
  if (alpha.value < 1) {
    const alphaHex = Math.round(alpha.value * 255).toString(16).padStart(2, '0')
    final = hex.length === 7 ? hex + alphaHex : hex.slice(0, 7) + alphaHex
  }
  emit('update:modelValue', final)
}

function handleHexInput(e: Event) {
  const val = (e.target as HTMLInputElement).value
  if (/^#[0-9a-fA-F]{3,8}$/.test(val) || /^[0-9a-fA-F]{3,8}$/.test(val)) {
    const hex = val.startsWith('#') ? val : '#' + val
    emit('update:modelValue', hex)
    parseColor(hex)
  }
}

function handleHexBlur(e: Event) {
  const val = (e.target as HTMLInputElement).value
  if (!/^#[0-9a-fA-F]{6}$/.test(val) && !/^#[0-9a-fA-F]{8}$/.test(val)) {
    (e.target as HTMLInputElement).value = props.modelValue
  }
}

let dragging: 'sv' | 'hue' | 'alpha' | null = null

function startSVDrag(e: MouseEvent) {
  dragging = 'sv'
  updateSV(e)
  document.addEventListener('mousemove', onSVMove)
  document.addEventListener('mouseup', stopDrag)
}

function onSVMove(e: MouseEvent) {
  if (dragging === 'sv') updateSV(e)
  else if (dragging === 'hue') updateHue(e)
  else if (dragging === 'alpha') updateAlpha(e)
}

function updateSV(e: MouseEvent) {
  if (!svPaletteRef.value) return
  const rect = svPaletteRef.value.getBoundingClientRect()
  svX.value = Math.max(0, Math.min(100, ((e.clientX - rect.left) / rect.width) * 100))
  svY.value = Math.max(0, Math.min(100, ((e.clientY - rect.top) / rect.height) * 100))
  emitColor(hexColor.value)
}

function startHueDrag(e: MouseEvent) {
  dragging = 'hue'
  updateHue(e)
  document.addEventListener('mousemove', onSVMove)
  document.addEventListener('mouseup', stopDrag)
}

function updateHue(e: MouseEvent) {
  if (!hueSliderRef.value) return
  const rect = hueSliderRef.value.getBoundingClientRect()
  hue.value = Math.max(0, Math.min(360, ((e.clientX - rect.left) / rect.width) * 360))
  emitColor(hexColor.value)
}

function startAlphaDrag(e: MouseEvent) {
  dragging = 'alpha'
  updateAlpha(e)
  document.addEventListener('mousemove', onSVMove)
  document.addEventListener('mouseup', stopDrag)
}

function updateAlpha(e: MouseEvent) {
  if (!alphaSliderRef.value) return
  const rect = alphaSliderRef.value.getBoundingClientRect()
  alpha.value = Math.max(0, Math.min(1, (e.clientX - rect.left) / rect.width))
  emitColor(hexColor.value)
}

function stopDrag() {
  dragging = null
  document.removeEventListener('mousemove', onSVMove)
  document.removeEventListener('mouseup', stopDrag)
}

function hexToRgb(hex: string) {
  const clean = hex.replace('#', '')
  if (clean.length < 6) return null
  return {
    r: parseInt(clean.slice(0, 2), 16),
    g: parseInt(clean.slice(2, 4), 16),
    b: parseInt(clean.slice(4, 6), 16),
  }
}

function rgbToHex(r: number, g: number, b: number) {
  return '#' + [r, g, b].map(x => Math.round(x).toString(16).padStart(2, '0')).join('')
}

function rgbToHsl(r: number, g: number, b: number) {
  r /= 255; g /= 255; b /= 255
  const max = Math.max(r, g, b), min = Math.min(r, g, b)
  let h = 0, s = 0
  const l = (max + min) / 2
  if (max !== min) {
    const d = max - min
    s = l > 0.5 ? d / (2 - max - min) : d / (max + min)
    switch (max) {
      case r: h = ((g - b) / d + (g < b ? 6 : 0)) / 6; break
      case g: h = ((b - r) / d + 2) / 6; break
      case b: h = ((r - g) / d + 4) / 6; break
    }
  }
  return { h: h * 360, s: s * 100, l: l * 100 }
}

function hslToRgb(h: number, s: number, l: number) {
  h /= 360; s /= 100; l /= 100
  let r = 0, g = 0, b = 0
  if (s === 0) {
    r = g = b = l
  } else {
    const hue2rgb = (p: number, q: number, t: number) => {
      if (t < 0) t += 1
      if (t > 1) t -= 1
      if (t < 1 / 6) return p + (q - p) * 6 * t
      if (t < 1 / 2) return q
      if (t < 2 / 3) return p + (q - p) * (2 / 3 - t) * 6
      return p
    }
    const q = l < 0.5 ? l * (1 + s) : l + s - l * s
    const p = 2 * l - q
    r = hue2rgb(p, q, h + 1 / 3)
    g = hue2rgb(p, q, h)
    b = hue2rgb(p, q, h - 1 / 3)
  }
  return { r: r * 255, g: g * 255, b: b * 255 }
}

function handleClickOutside(e: MouseEvent) {
  if (!open.value) return
  const target = e.target as HTMLElement
  if (containerRef.value?.contains(target)) return
  if (dropdownRef.value?.contains(target)) return
  open.value = false
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  window.addEventListener('resize', updateDropdownPosition)
  window.addEventListener('scroll', updateDropdownPosition, true)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
  window.removeEventListener('resize', updateDropdownPosition)
  window.removeEventListener('scroll', updateDropdownPosition, true)
  document.removeEventListener('mousemove', onSVMove)
  document.removeEventListener('mouseup', stopDrag)
})
</script>

<style scoped>
.color-picker-popup {
  position: relative;
  display: inline-flex;
}

.color-trigger {
  width: 72px;
  height: 36px;
  border-radius: 6px;
  border: 2px solid rgba(0, 0, 0, 0.08);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  padding: 0;
}
.color-trigger:hover {
  transform: scale(1.05);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}
.color-trigger-text {
  font-size: 11px;
  font-family: monospace;
  color: #fff;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.3);
  mix-blend-mode: difference;
}

.color-picker-backdrop {
  position: fixed;
  inset: 0;
  z-index: 10001;
}

.color-picker-dropdown {
  width: 260px;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.15), 0 2px 8px rgba(0, 0, 0, 0.08);
  border: 1px solid rgba(0, 0, 0, 0.08);
}

.color-picker-body {
  padding: 12px;
}

.sv-palette {
  position: relative;
  width: 100%;
  height: 140px;
  border-radius: 6px;
  overflow: hidden;
  cursor: crosshair;
  margin-bottom: 10px;
}
.sv-palette-white,
.sv-palette-black {
  position: absolute;
  inset: 0;
}
.sv-handle {
  position: absolute;
  width: 14px;
  height: 14px;
  border-radius: 50%;
  border: 2px solid #fff;
  box-shadow: 0 0 4px rgba(0, 0, 0, 0.3), inset 0 0 2px rgba(0, 0, 0, 0.2);
  transform: translate(-50%, -50%);
  pointer-events: none;
}

.sliders-wrap {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 10px;
}

.hue-slider {
  position: relative;
  height: 12px;
  border-radius: 6px;
  background: linear-gradient(90deg, red, #ff0 16.66%, #0f0 33.33%, #0ff 50%, #00f 66.66%, #f0f 83.33%, red);
  cursor: pointer;
}
.hue-handle {
  position: absolute;
  top: 50%;
  width: 14px;
  height: 14px;
  border-radius: 50%;
  border: 2px solid #fff;
  box-shadow: 0 0 4px rgba(0, 0, 0, 0.3);
  transform: translate(-50%, -50%);
  pointer-events: none;
  background: transparent;
}

.alpha-slider {
  position: relative;
  height: 12px;
  border-radius: 6px;
  cursor: pointer;
  background-image: linear-gradient(45deg, #ccc 25%, transparent 25%),
    linear-gradient(-45deg, #ccc 25%, transparent 25%),
    linear-gradient(45deg, transparent 75%, #ccc 75%),
    linear-gradient(-45deg, transparent 75%, #ccc 75%);
  background-size: 8px 8px;
  background-position: 0 0, 0 4px, 4px -4px, -4px 0;
}
.alpha-slider-bg {
  position: absolute;
  inset: 0;
  border-radius: 6px;
}
.alpha-handle {
  position: absolute;
  top: 50%;
  width: 14px;
  height: 14px;
  border-radius: 50%;
  border: 2px solid #fff;
  box-shadow: 0 0 4px rgba(0, 0, 0, 0.3);
  transform: translate(-50%, -50%);
  pointer-events: none;
  background: transparent;
}

.preset-colors {
  display: grid;
  grid-template-columns: repeat(10, 1fr);
  gap: 4px;
  margin-bottom: 10px;
}
.preset-dot {
  width: 100%;
  aspect-ratio: 1;
  border-radius: 4px;
  border: 1.5px solid transparent;
  cursor: pointer;
  transition: all 0.15s;
  padding: 0;
}
.preset-dot:hover {
  transform: scale(1.2);
  z-index: 1;
}
.preset-dot.active {
  border-color: #333;
  box-shadow: 0 0 0 1px #fff, 0 0 0 3px currentColor;
}

.hex-input-row {
  display: flex;
  align-items: center;
  gap: 8px;
}
.hex-label {
  font-size: 12px;
  color: #999;
  font-weight: 500;
}
.hex-input {
  flex: 1;
  height: 28px;
  border: 1px solid #ddd;
  border-radius: 4px;
  padding: 0 8px;
  font-size: 13px;
  font-family: monospace;
  outline: none;
  color: #333;
}
.hex-input:focus {
  border-color: var(--primary, #2d8cf0);
}
</style>
