<template>
  <div class="icon-picker">
    <!-- 触发按钮 -->
    <el-input
      :model-value="modelValue"
      placeholder="点击选择图标"
      readonly
      @click="openPicker"
      class="icon-input"
    >
      <template #prefix>
        <span class="icon-preview" v-if="modelValue">
          <template v-if="isEmoji(modelValue)">{{ modelValue }}</template>
          <el-icon v-else><component :is="modelValue" /></el-icon>
        </span>
      </template>
      <template #append>
        <el-button :icon="Grid" @click="openPicker" />
      </template>
    </el-input>

    <!-- 图标选择弹窗 -->
    <el-dialog
      v-model="pickerVisible"
      title="选择图标"
      width="720px"
      :close-on-click-modal="true"
      draggable
      append-to-body
    >
      <!-- 搜索 + 分类 -->
      <div class="picker-toolbar">
        <el-input
          v-model="searchKeyword"
          placeholder="搜索图标名称..."
          :prefix-icon="Search"
          clearable
          style="width: 240px"
        />
        <el-radio-group v-model="activeCategory" size="small">
          <el-radio-button label="all">全部</el-radio-button>
          <el-radio-button label="emoji">Emoji</el-radio-button>
          <el-radio-button label="ep">Element Plus</el-radio-button>
        </el-radio-group>
      </div>

      <!-- 当前选中 -->
      <div class="current-icon" v-if="modelValue">
        <span class="label">当前选择：</span>
        <span class="value">
          <template v-if="isEmoji(modelValue)">{{ modelValue }}</template>
          <el-icon v-else><component :is="modelValue" /></el-icon>
          {{ modelValue }}
        </span>
        <el-button text type="danger" size="small" @click="clearIcon">清除</el-button>
      </div>

      <!-- Emoji 区 -->
      <div v-if="activeCategory === 'all' || activeCategory === 'emoji'" class="icon-section">
        <div class="section-title">常用 Emoji</div>
        <div class="icon-grid">
          <div
            v-for="emoji in filteredEmojis"
            :key="emoji"
            class="icon-cell"
            :class="{ active: modelValue === emoji }"
            @click="selectIcon(emoji)"
          >
            {{ emoji }}
          </div>
        </div>
      </div>

      <!-- Element Plus 图标区 -->
      <div v-if="activeCategory === 'all' || activeCategory === 'ep'" class="icon-section">
        <div class="section-title">Element Plus 图标 ({{ filteredEpIcons.length }})</div>
        <div class="icon-grid ep-grid">
          <div
            v-for="name in filteredEpIcons"
            :key="name"
            class="icon-cell ep-cell"
            :class="{ active: modelValue === name }"
            @click="selectIcon(name)"
            :title="name"
          >
            <el-icon :size="18"><component :is="name" /></el-icon>
            <span class="ep-name">{{ name }}</span>
          </div>
        </div>
      </div>

      <template #footer>
        <el-button @click="pickerVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { Search, Grid } from '@element-plus/icons-vue'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'

const props = defineProps<{
  modelValue: string
}>()
const emit = defineEmits(['update:modelValue'])

const pickerVisible = ref(false)
const searchKeyword = ref('')
const activeCategory = ref('all')

/** 常用 Emoji 列表 */
const emojiList = [
  '🏭', '☀️', '⚡', '💧', '🔋', '🌐', '📊', '📡', '🔌', '⚙️',
  '🏠', '🏢', '🏗️', '🚂', '🚢', '✈️', '🚗', '🛠️', '🔧', '🔨',
  '🌲', '🍃', '🌍', '♻️', '🌱', '🔥', '💨', '🌪️', '🌊', '🏔️',
  '🌾', '🐮', '🐔', '🐟', '🚜', '🏭', '⚗️', '🔬', '🧪', '🩺',
  '🚿', '🚰', '🛢️', '⚖️', '📈', '📉', '🎛️', '🧮', '💡', '🔦'
]

/** Element Plus 图标名列表 */
const epIconNames = Object.keys(ElementPlusIconsVue)

/** 判断是否为 Emoji */
function isEmoji(str: string): boolean {
  // Emoji 的 Unicode 范围检测
  return /[\u{1F000}-\u{1FFFF}]|[\u{2600}-\u{27BF}]/u.test(str)
}

/** 过滤后的 Emoji */
const filteredEmojis = computed(() => {
  if (!searchKeyword.value) return emojiList
  return emojiList.filter(() => true) // Emoji 无法按名称搜索，直接返回全部
})

/** 过滤后的 EP 图标 */
const filteredEpIcons = computed(() => {
  if (!searchKeyword.value) return epIconNames
  const kw = searchKeyword.value.toLowerCase()
  return epIconNames.filter(name => name.toLowerCase().includes(kw))
})

function openPicker() {
  pickerVisible.value = true
}

function selectIcon(icon: string) {
  emit('update:modelValue', icon)
  pickerVisible.value = false
}

function clearIcon() {
  emit('update:modelValue', '')
}
</script>

<style scoped lang="scss">
.icon-picker {
  .icon-input {
    width: 220px;
    cursor: pointer;
  }

  .icon-preview {
    font-size: 18px;
    display: inline-flex;
    align-items: center;
  }
}

.picker-toolbar {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 16px;
}

.current-icon {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: rgba(75, 63, 227, 0.08);
  border-radius: 6px;
  margin-bottom: 16px;
  font-size: 14px;

  .label {
    color: #8b949e;
  }
  .value {
    color: #e6edf3;
    font-weight: 600;
    display: flex;
    align-items: center;
    gap: 4px;
  }
}

.icon-section {
  margin-bottom: 20px;

  .section-title {
    font-size: 12px;
    color: #6e7681;
    margin-bottom: 8px;
    text-transform: uppercase;
    letter-spacing: 1px;
  }
}

.icon-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(44px, 1fr));
  gap: 6px;
  max-height: 200px;
  overflow-y: auto;
  padding: 4px;
  border: 1px solid #30363d;
  border-radius: 6px;
}

.ep-grid {
  grid-template-columns: repeat(auto-fill, minmax(80px, 1fr));
  max-height: 300px;
}

.icon-cell {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 40px;
  font-size: 20px;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.15s;

  &:hover {
    background: rgba(75, 63, 227, 0.12);
  }

  &.active {
    background: rgba(75, 63, 227, 0.2);
    box-shadow: 0 0 0 1px #4b3fe3;
  }
}

.ep-cell {
  flex-direction: column;
  height: 56px;
  gap: 2px;

  .ep-name {
    font-size: 9px;
    color: #6e7681;
    text-align: center;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    max-width: 72px;
  }

  &:hover .ep-name {
    color: #c9d1d9;
  }
}
</style>
