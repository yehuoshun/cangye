<template>
  <div class="tag-input-wrapper" ref="wrapperRef" @click="focusInput">
    <span
      v-for="(tag, i) in selectedTags"
      :key="tag.id"
      class="tag-pill"
      :class="tag.color === 'gray' ? 'default' : tag.color"
      :style="{ background: tag.color && tag.color !== 'gray' ? tag.color + '33' : undefined }"
    >
      {{ tag.name }}
      <span class="remove" @click.stop="removeTag(i)">✕</span>
    </span>
    <input
      ref="inputRef"
      v-model="query"
      :placeholder="placeholder"
      @input="onInput"
      @keydown.enter="onEnter"
      @keydown.backspace="onBackspace"
      @focus="showDropdown = true"
      @blur="hideDropdown"
    />
    <div class="tag-dropdown" v-if="showDropdown && suggestions.length > 0">
      <div
        v-for="tag in suggestions"
        :key="tag.id"
        class="tag-dropdown-item"
        @mousedown.prevent="selectTag(tag)"
      >
        {{ tag.name }}
      </div>
      <div
        v-if="query && suggestions.length === 0"
        class="tag-dropdown-item"
        @mousedown.prevent="createTag"
        style="color: var(--text-accent)"
      >
        + 创建"{{ query }}"
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { api } from '@/api'
import type { Tag } from '@/api'

const props = defineProps<{
  modelValue: Tag[]
  placeholder?: string
}>()

const emit = defineEmits<{
  'update:modelValue': [tags: Tag[]]
}>()

const selectedTags = ref<Tag[]>([...props.modelValue])
const query = ref('')
const suggestions = ref<Tag[]>([])
const showDropdown = ref(false)
const inputRef = ref<HTMLInputElement | null>(null)
const wrapperRef = ref<HTMLDivElement | null>(null)

let debounceTimer: ReturnType<typeof setTimeout> | null = null

watch(() => props.modelValue, (v) => {
  selectedTags.value = [...v]
})

function focusInput() {
  inputRef.value?.focus()
}

function onInput() {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(async () => {
    if (!query.value.trim()) {
      suggestions.value = []
      return
    }
    try {
      suggestions.value = await api.tags.search(query.value)
      showDropdown.value = true
    } catch {
      suggestions.value = []
    }
  }, 300)
}

function onEnter() {
  if (query.value.trim()) {
    // If there are suggestions, select first
    if (suggestions.value.length > 0) {
      selectTag(suggestions.value[0])
    } else {
      createTag()
    }
  }
}

function onBackspace() {
  if (!query.value && selectedTags.value.length > 0) {
    selectedTags.value.pop()
    emit('update:modelValue', [...selectedTags.value])
  }
}

function selectTag(tag: Tag) {
  if (!selectedTags.value.find(t => t.id === tag.id)) {
    selectedTags.value.push(tag)
    emit('update:modelValue', [...selectedTags.value])
  }
  query.value = ''
  suggestions.value = []
}

async function createTag() {
  try {
    const tag = await api.tags.create({ name: query.value.trim() })
    selectedTags.value.push(tag)
    emit('update:modelValue', [...selectedTags.value])
    query.value = ''
    suggestions.value = []
  } catch (e: any) {
    console.error('create tag failed', e)
  }
}

function removeTag(index: number) {
  selectedTags.value.splice(index, 1)
  emit('update:modelValue', [...selectedTags.value])
}

function hideDropdown() {
  setTimeout(() => { showDropdown.value = false }, 200)
}
</script>
