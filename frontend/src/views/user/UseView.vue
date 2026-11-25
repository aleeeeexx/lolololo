<template>
  <AppLayout>
    <div class="space-y-5">
      <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
        <div>
          <h1 class="text-xl font-semibold text-gray-900 dark:text-gray-100">{{ t('playground.title') }}</h1>
          <p class="mt-1 max-w-3xl text-sm text-[color:var(--muted)]">{{ t('playground.description') }}</p>
        </div>
        <button type="button" class="btn btn-secondary" :disabled="sessionLoading" @click="loadSession">
          <Icon name="refresh" size="md" :class="sessionLoading ? 'animate-spin' : ''" />
          {{ t('playground.refreshSession') }}
        </button>
      </div>

      <div
        class="flex items-start gap-3 rounded-lg border border-sky-200 bg-sky-50 px-3 py-3 text-sm leading-6 text-sky-900 dark:border-sky-800/50 dark:bg-sky-950/30 dark:text-sky-100"
        data-testid="playground-client-tip"
        role="note"
      >
        <Icon name="infoCircle" size="md" class="mt-0.5 flex-shrink-0 text-sky-500 dark:text-sky-300" />
        <p>
          <i18n-t keypath="playground.clientTip" tag="span" scope="global">
            <template #handbookLink>
              <router-link
                :to="{ name: 'Handbook', hash: '#quick-examples' }"
                class="font-medium underline underline-offset-2 hover:text-sky-700 dark:hover:text-white"
              >
                {{ t('playground.clientTipLink') }}
              </router-link>
            </template>
          </i18n-t>
        </p>
      </div>

      <div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
        <div class="card p-4">
          <p class="text-xs font-medium text-[color:var(--muted)]">{{ t('playground.remainingBalance') }}</p>
          <p class="mt-1 text-2xl font-semibold text-emerald-600 dark:text-emerald-400">
            {{ formatCurrency(session?.available_balance ?? 0) }}
          </p>
          <p v-if="(session?.frozen_balance ?? 0) > 0" class="mt-1 text-xs text-[color:var(--muted)]">
            {{ t('playground.frozenBalance') }}: {{ formatCurrency(session?.frozen_balance ?? 0) }}
          </p>
        </div>
        <div class="card p-4">
          <p class="text-xs font-medium text-[color:var(--muted)]">{{ t('playground.lastCost') }}</p>
          <p class="mt-1 text-2xl font-semibold text-gray-900 dark:text-gray-100">
            {{ lastCostLabel }}
          </p>
          <p v-if="costStatus" class="mt-1 text-xs text-[color:var(--muted)]">{{ costStatus }}</p>
        </div>
        <div v-if="session?.subscriptions?.length" class="card p-4 sm:col-span-2 xl:col-span-1">
          <p class="text-xs font-medium text-[color:var(--muted)]">{{ t('playground.subscriptionQuota') }}</p>
          <ul class="mt-2 space-y-1 text-sm text-gray-700 dark:text-gray-300">
            <li v-for="item in session.subscriptions" :key="item.id">
              <span class="font-medium">{{ item.group_name || item.group_id }}</span>
              <span class="ml-2 text-xs text-[color:var(--muted)]">
                {{ t('playground.subscriptionDaily') }} {{ formatCurrency(item.daily_used_usd) }}/{{ formatLimit(item.daily_limit_usd) }}
                · {{ t('playground.subscriptionMonthly') }} {{ formatCurrency(item.monthly_used_usd) }}/{{ formatLimit(item.monthly_limit_usd) }}
              </span>
            </li>
          </ul>
        </div>
      </div>

      <div v-if="session && !boundGroups.length" class="card p-4" data-testid="playground-need-key">
        <p class="text-sm text-gray-700 dark:text-gray-300">{{ t('playground.needPersonalKey') }}</p>
        <router-link :to="{ name: 'Keys' }" class="btn btn-secondary mt-3 inline-flex">
          {{ t('playground.goToKeys') }}
        </router-link>
      </div>

      <div class="flex gap-2" role="tablist" :aria-label="t('playground.title')">
        <button
          type="button"
          role="tab"
          class="supported-model-pill"
          :class="tab === 'chat' ? 'supported-model-pill-active' : ''"
          :aria-selected="tab === 'chat'"
          data-testid="playground-tab-chat"
          @click="tab = 'chat'"
        >
          {{ t('playground.tabChat') }}
        </button>
        <button
          type="button"
          role="tab"
          class="supported-model-pill"
          :class="tab === 'image' ? 'supported-model-pill-active' : ''"
          :aria-selected="tab === 'image'"
          data-testid="playground-tab-image"
          @click="tab = 'image'"
        >
          {{ t('playground.tabImage') }}
        </button>
      </div>

      <div v-if="tab === 'chat'" class="grid gap-4 lg:grid-cols-[minmax(0,1fr)_16rem]">
        <div class="card flex min-h-[32rem] flex-col overflow-hidden">
          <div class="flex flex-wrap items-end gap-3 border-b border-gray-100 p-4 dark:border-gray-800">
            <div class="min-w-[14rem] flex-1">
              <label class="input-label">{{ t('playground.group') }}</label>
              <Select v-model="selectedGroupId" :options="groupOptions" :placeholder="t('playground.selectGroup')" />
            </div>
            <div class="min-w-[16rem] flex-1">
              <label class="input-label">{{ t('playground.model') }}</label>
              <Select v-model="chatModel" :options="chatModelOptions" :placeholder="t('playground.selectChatModel')" />
            </div>
            <button type="button" class="btn btn-secondary" @click="startNewChat">{{ t('playground.newConversation') }}</button>
          </div>

          <div v-if="!chatMessages.length" class="flex flex-1 items-center justify-center px-6 text-sm text-[color:var(--muted)]">
            {{ chatEmptyLabel }}
          </div>
          <div v-else ref="transcriptRef" class="flex-1 space-y-4 overflow-y-auto p-4">
            <div v-for="(message, index) in chatMessages" :key="index" class="flex" :class="message.role === 'user' ? 'justify-end' : 'justify-start'">
              <div
                class="max-w-[85%] whitespace-pre-wrap rounded-2xl px-4 py-3 text-sm leading-6"
                :class="message.role === 'user'
                  ? 'bg-sky-600 text-white'
                  : 'bg-gray-100 text-gray-900 dark:bg-gray-800 dark:text-gray-100'"
              >
                <p class="mb-1 text-xs opacity-70">{{ message.role === 'user' ? t('playground.you') : t('playground.assistant') }}</p>
                {{ message.content }}
              </div>
            </div>
          </div>

          <div class="border-t border-gray-100 p-4 dark:border-gray-800">
            <div v-if="attachments.length" class="mb-3 flex flex-wrap gap-2">
              <span
                v-for="file in attachments"
                :key="file.filename"
                class="inline-flex items-center gap-2 rounded-full bg-gray-100 px-3 py-1 text-xs text-gray-700 dark:bg-gray-800 dark:text-gray-200"
              >
                {{ file.filename }}
                <button type="button" class="text-[color:var(--muted)]" :aria-label="t('playground.removeAttachment')" @click="removeAttachment(file.filename)">×</button>
              </span>
            </div>
            <p class="mb-2 text-xs text-[color:var(--muted)]">
              {{ t('playground.documentLimit', { count: maxDocuments, size: formatBytes(maxDocumentBytes) }) }}
            </p>
            <div class="flex flex-col gap-3 sm:flex-row sm:items-end">
              <textarea
                v-model="draft"
                class="input min-h-[5.5rem] flex-1 resize-y"
                :placeholder="t('playground.composerPlaceholder')"
                :disabled="chatSending"
                @keydown.enter.exact.prevent="sendChat"
              />
              <div class="flex flex-wrap gap-2">
                <label class="btn btn-secondary cursor-pointer">
                  {{ t('playground.attach') }}
                  <input type="file" class="hidden" accept=".pdf,.txt,.md,text/plain,application/pdf" multiple @change="onAttachFiles">
                </label>
                <button
                  v-if="chatSending"
                  type="button"
                  class="btn btn-secondary"
                  @click="stopChat"
                >
                  {{ t('playground.stop') }}
                </button>
                <button
                  v-else
                  type="button"
                  class="btn btn-primary"
                  :disabled="!canSend"
                  data-testid="playground-send"
                  @click="sendChat"
                >
                  {{ t('playground.send') }}
                </button>
              </div>
            </div>
            <p v-if="chatError" class="mt-2 text-sm text-red-600 dark:text-red-400">{{ chatError }}</p>
          </div>
        </div>

        <div class="card p-4">
          <h2 class="text-sm font-medium text-gray-900 dark:text-gray-100">{{ t('playground.conversations') }}</h2>
          <p v-if="!conversations.length" class="mt-3 text-sm text-[color:var(--muted)]">{{ t('playground.noConversations') }}</p>
          <ul v-else class="mt-3 space-y-2">
            <li v-for="item in conversations" :key="item.id">
              <button
                type="button"
                class="w-full rounded-lg px-3 py-2 text-left text-sm hover:bg-gray-50 dark:hover:bg-gray-800"
                :class="item.id === activeConversationId ? 'bg-gray-100 dark:bg-gray-800' : ''"
                @click="openConversation(item.id)"
              >
                <span class="block truncate font-medium text-gray-800 dark:text-gray-100">{{ item.title }}</span>
                <span class="block truncate text-xs text-[color:var(--muted)]">{{ item.model }}</span>
              </button>
              <button type="button" class="mt-1 text-xs text-[color:var(--muted)]" @click="deleteConversation(item.id)">
                {{ t('playground.deleteConversation') }}
              </button>
            </li>
          </ul>
        </div>
      </div>

      <div v-else class="card p-4 sm:p-6">
        <div class="grid gap-4 lg:grid-cols-[minmax(0,1fr)_20rem]">
          <div class="space-y-4">
            <div>
              <label class="input-label">{{ t('playground.group') }}</label>
              <Select v-model="selectedGroupId" :options="groupOptions" :placeholder="t('playground.selectGroup')" />
            </div>
            <div>
              <label class="input-label">{{ t('playground.model') }}</label>
              <Select v-model="imageModel" :options="imageModelOptions" :placeholder="t('playground.selectImageModel')" />
            </div>
            <div>
              <label class="input-label">{{ t('playground.imagePrompt') }}</label>
              <textarea
                v-model="imagePrompt"
                class="input min-h-[8rem] resize-y"
                :placeholder="t('playground.imagePromptPlaceholder')"
                :disabled="imageBusy"
              />
            </div>
            <button type="button" class="btn btn-primary" :disabled="!canGenerateImage" data-testid="playground-generate" @click="generateImage">
              {{ imageBusy ? imageStatusLabel : t('playground.generate') }}
            </button>
            <p v-if="imageError" class="text-sm text-red-600 dark:text-red-400">{{ imageError }}</p>
          </div>
          <div>
            <div v-if="!imageOutputs.length" class="flex min-h-[16rem] items-center justify-center rounded-xl border border-dashed border-gray-200 px-4 text-center text-sm text-[color:var(--muted)] dark:border-gray-700">
              {{ imageEmptyLabel }}
            </div>
            <div v-else class="space-y-4">
              <div v-for="(output, index) in imageOutputs" :key="index" class="overflow-hidden rounded-xl border border-gray-200 dark:border-gray-700">
                <img v-if="previewSrc(output)" :src="previewSrc(output)" alt="" class="w-full bg-gray-50 object-contain dark:bg-gray-900">
                <div class="flex items-center justify-between gap-3 p-3">
                  <span class="text-sm text-gray-700 dark:text-gray-200">{{ t('playground.imageReady') }} #{{ index + 1 }}</span>
                  <button type="button" class="btn btn-secondary btn-sm" data-testid="playground-download" @click="downloadOutput(output, index)">
                    {{ t('playground.download') }}
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import Select, { type SelectOption } from '@/components/common/Select.vue'
import {
  fetchPlaygroundImageContent,
  getPlaygroundImageTask,
  getPlaygroundSession,
  pollPlaygroundRequestCost,
  submitPlaygroundChat,
  submitPlaygroundImage,
  type PlaygroundDocumentPayload,
  type PlaygroundImageResult,
  type PlaygroundSession,
} from '@/api/playground'
import { useAuthStore } from '@/stores/auth'
import { formatBytes, formatCurrency } from '@/utils/format'
import { downloadPlaygroundImage, type PlaygroundImageSource } from '@/utils/playgroundDownload'
import {
  loadPlaygroundHistory,
  savePlaygroundHistory,
  upsertPlaygroundConversation,
  type PlaygroundHistoryConversation,
  type PlaygroundHistoryMessage,
} from '@/utils/playgroundHistory'
import {
  appendPlaygroundAttachmentText,
  modelsForPlaygroundGroup,
  pickDefaultPlaygroundGroup,
  retainOrPickPlaygroundModel,
} from '@/utils/playgroundGroups'
import { formatMultiplier } from '@/utils/formatters'

const { t } = useI18n()
const authStore = useAuthStore()

const tab = ref<'chat' | 'image'>('chat')
const session = ref<PlaygroundSession | null>(null)
const sessionLoading = ref(false)
const lastCost = ref<number | null>(null)
const costStatus = ref('')
const chatModel = ref('')
const imageModel = ref('')
const selectedGroupId = ref<number | null>(null)
const draft = ref('')
const attachments = ref<PlaygroundDocumentPayload[]>([])
const chatMessages = ref<PlaygroundHistoryMessage[]>([])
const conversations = ref<PlaygroundHistoryConversation[]>([])
const activeConversationId = ref('')
const chatSending = ref(false)
const chatError = ref('')
const imagePrompt = ref('')
const imageBusy = ref(false)
const imageError = ref('')
const imageStatus = ref<'idle' | 'submitting' | 'polling'>('idle')
const imageTaskId = ref('')
const imageOutputs = ref<PlaygroundImageSource[]>([])
const transcriptRef = ref<HTMLElement | null>(null)
let chatAbort: AbortController | null = null
let pollTimer: number | null = null

const userId = computed(() => authStore.user?.id ?? 0)
const maxDocuments = computed(() => session.value?.limits.max_documents ?? 3)
const maxDocumentBytes = computed(() => session.value?.limits.max_document_bytes ?? 5 * 1024 * 1024)
const boundGroups = computed(() => session.value?.groups ?? [])
const selectedGroup = computed(() => boundGroups.value.find((item) => item.id === selectedGroupId.value))
const chatModels = computed(() => modelsForPlaygroundGroup(selectedGroup.value, 'chat'))
const imageModels = computed(() => modelsForPlaygroundGroup(selectedGroup.value, 'image'))
const chatModelOptions = computed<SelectOption[]>(() => chatModels.value.map((name) => ({ value: name, label: name })))
const imageModelOptions = computed<SelectOption[]>(() => imageModels.value.map((name) => ({ value: name, label: name })))
const groupOptions = computed<SelectOption[]>(() => boundGroups.value.map((item) => ({
  value: item.id,
  label: groupLabel(item),
})))
const canSend = computed(() => !chatSending.value && !!chatModel.value && selectedGroupId.value != null && (!!draft.value.trim() || attachments.value.length > 0))
const canGenerateImage = computed(() => !imageBusy.value && !!imageModel.value && selectedGroupId.value != null && !!imagePrompt.value.trim())
const chatEmptyLabel = computed(() => {
  if (!boundGroups.value.length) return t('playground.needPersonalKey')
  if (!chatModels.value.length) return t('playground.noChatModels')
  return t('playground.emptyChat')
})
const imageEmptyLabel = computed(() => {
  if (!boundGroups.value.length) return t('playground.needPersonalKey')
  if (!imageModels.value.length) return t('playground.noImageModels')
  return t('playground.imageEmpty')
})
const imageStatusLabel = computed(() => imageStatus.value === 'polling' ? t('playground.polling') : t('playground.generating'))
const lastCostLabel = computed(() => lastCost.value == null ? '—' : formatCurrency(lastCost.value))

function formatLimit(value: number): string {
  return value > 0 ? formatCurrency(value) : '∞'
}

function groupLabel(item: { name: string; subscription_type: boolean; has_remaining_quota: boolean; rate_multiplier: number }): string {
  const kind = item.subscription_type ? t('playground.groupSubscription') : t('playground.groupPayAsYouGo')
  const quota = item.subscription_type
    ? (item.has_remaining_quota ? t('playground.groupQuotaLeft') : t('playground.groupQuotaExhausted'))
    : ''
  const rate = t('playground.groupRate', { rate: formatMultiplier(item.rate_multiplier) })
  return [item.name, kind, quota, rate].filter(Boolean).join(' · ')
}

function syncSelectedGroup() {
  const groups = boundGroups.value
  if (!groups.length) {
    selectedGroupId.value = null
    return
  }
  if (selectedGroupId.value != null && groups.some((item) => item.id === selectedGroupId.value)) {
    return
  }
  selectedGroupId.value = pickDefaultPlaygroundGroup(groups)?.id ?? groups[0].id
}

function syncSelectedModels() {
  chatModel.value = retainOrPickPlaygroundModel(chatModels.value, chatModel.value)
  imageModel.value = retainOrPickPlaygroundModel(imageModels.value, imageModel.value)
}

async function loadSession() {
  sessionLoading.value = true
  try {
    session.value = await getPlaygroundSession()
    syncSelectedGroup()
    syncSelectedModels()
  } catch (error) {
    chatError.value = chatError.value || t('playground.loadSessionFailed')
    console.error(error)
  } finally {
    sessionLoading.value = false
  }
}

function persistHistory() {
  if (!userId.value) return
  savePlaygroundHistory(userId.value, conversations.value)
}

function startNewChat() {
  activeConversationId.value = ''
  chatMessages.value = []
  draft.value = ''
  attachments.value = []
  chatError.value = ''
}

function openConversation(id: string) {
  const item = conversations.value.find((conversation) => conversation.id === id)
  if (!item) return
  activeConversationId.value = item.id
  chatModel.value = item.model || chatModel.value
  chatMessages.value = [...item.messages]
}

function deleteConversation(id: string) {
  conversations.value = conversations.value.filter((item) => item.id !== id)
  if (activeConversationId.value === id) startNewChat()
  persistHistory()
}

function upsertCurrentConversation(titleHint: string) {
  const id = activeConversationId.value || `c_${Date.now()}`
  activeConversationId.value = id
  const title = (chatMessages.value.find((item) => item.role === 'user')?.content || titleHint || 'Chat').trim().slice(0, 80)
  conversations.value = upsertPlaygroundConversation(conversations.value, {
    id,
    title,
    model: chatModel.value,
    messages: [...chatMessages.value],
    updatedAt: Date.now(),
  })
  persistHistory()
}

async function onAttachFiles(event: Event) {
  const input = event.target as HTMLInputElement
  const files = Array.from(input.files || [])
  input.value = ''
  for (const file of files) {
    if (attachments.value.length >= maxDocuments.value) break
    if (file.size > maxDocumentBytes.value) {
      chatError.value = t('playground.documentLimit', { count: maxDocuments.value, size: formatBytes(maxDocumentBytes.value) })
      continue
    }
    const payload = await fileToDocument(file)
    if (payload) attachments.value.push(payload)
  }
}

async function fileToDocument(file: File): Promise<PlaygroundDocumentPayload | null> {
  const name = file.name || 'document'
  const lower = name.toLowerCase()
  if (lower.endsWith('.txt') || lower.endsWith('.md') || file.type.startsWith('text/')) {
    return { filename: name, text: await file.text() }
  }
  if (lower.endsWith('.pdf') || file.type === 'application/pdf') {
    const buffer = await file.arrayBuffer()
    const bytes = new Uint8Array(buffer)
    let binary = ''
    bytes.forEach((value) => {
      binary += String.fromCharCode(value)
    })
    return { filename: name, data: btoa(binary) }
  }
  chatError.value = t('playground.documentLimit', { count: maxDocuments.value, size: formatBytes(maxDocumentBytes.value) })
  return null
}

function removeAttachment(name: string) {
  attachments.value = attachments.value.filter((item) => item.filename !== name)
}

async function scrollTranscript() {
  await nextTick()
  if (transcriptRef.value) {
    transcriptRef.value.scrollTop = transcriptRef.value.scrollHeight
  }
}

function stopChat() {
  chatAbort?.abort()
  chatSending.value = false
}

async function sendChat() {
  if (!canSend.value) return
  const content = draft.value.trim()
  const documents = [...attachments.value]
  draft.value = ''
  attachments.value = []
  chatError.value = ''
  if (content) {
    chatMessages.value.push({ role: 'user', content })
  } else if (documents.length) {
    chatMessages.value.push({ role: 'user', content: documents.map((item) => item.filename).join(', ') })
  }
  const assistantIndex = chatMessages.value.length
  chatMessages.value.push({ role: 'assistant', content: '' })
  await scrollTranscript()
  chatSending.value = true
  chatAbort = new AbortController()
  try {
    const history = chatMessages.value
      .slice(0, assistantIndex)
      .map((item) => ({ role: item.role, content: item.content }))
    const result = await submitPlaygroundChat({
      model: chatModel.value,
      groupId: selectedGroupId.value as number,
      messages: history,
      documents,
      signal: chatAbort.signal,
      onDelta: async (delta) => {
        const current = chatMessages.value[assistantIndex]
        if (current) current.content += delta
        chatMessages.value = [...chatMessages.value]
        await scrollTranscript()
      },
    })
    if (!chatMessages.value[assistantIndex]?.content) {
      chatMessages.value[assistantIndex].content = result.text
    }
    const userIndex = assistantIndex - 1
    if (documents.length && chatMessages.value[userIndex]) {
      chatMessages.value[userIndex].content = appendPlaygroundAttachmentText(
        chatMessages.value[userIndex].content,
        documents
      )
    }
    upsertCurrentConversation(content)
    if (result.requestId) await refreshCost(result.requestId)
    await loadSession()
  } catch (error) {
    if ((error as Error).name === 'AbortError') return
    chatError.value = error instanceof Error && error.message ? error.message : t('playground.error')
    if (!chatMessages.value[assistantIndex]?.content) {
      chatMessages.value.splice(assistantIndex, 1)
    }
  } finally {
    chatSending.value = false
    chatAbort = null
  }
}

async function refreshCost(requestId: string) {
  costStatus.value = t('playground.costPending')
  try {
    const cost = await pollPlaygroundRequestCost(requestId)
    if (cost.found) {
      lastCost.value = cost.actual_cost
      costStatus.value = ''
      return
    }
    costStatus.value = t('playground.costUnavailable')
  } catch {
    costStatus.value = t('playground.costUnavailable')
  }
}

async function generateImage() {
  if (!canGenerateImage.value) return
  imageError.value = ''
  imageBusy.value = true
  imageStatus.value = 'submitting'
  imageOutputs.value = []
  try {
    const submitted = await submitPlaygroundImage({
      model: imageModel.value,
      groupId: selectedGroupId.value as number,
      prompt: imagePrompt.value.trim(),
    })
    const taskId = submitted.task.task_id || submitted.task.id
    imageTaskId.value = taskId
    imageStatus.value = 'polling'
    const completed = await pollImageTask(taskId)
    const data = (completed.result as PlaygroundImageResult | undefined)?.data || []
    imageOutputs.value = data.map((item) => ({ url: item.url, b64_json: item.b64_json }))
    if (!imageOutputs.value.length && completed.image_url) {
      imageOutputs.value = [{ url: completed.image_url }]
    }
    if (submitted.requestId) await refreshCost(submitted.requestId)
    await loadSession()
  } catch (error) {
    imageError.value = error instanceof Error && error.message ? error.message : t('playground.error')
  } finally {
    imageBusy.value = false
    imageStatus.value = 'idle'
  }
}

async function pollImageTask(taskId: string) {
  for (let i = 0; i < 80; i++) {
    const task = await getPlaygroundImageTask(taskId)
    if (task.status === 'completed') return task
    if (task.status === 'failed') {
      const message = typeof task.error === 'string' ? task.error : task.error?.message
      throw new Error(message || t('playground.error'))
    }
    await new Promise((resolve) => {
      pollTimer = window.setTimeout(resolve, 3000)
    })
  }
  throw new Error(t('playground.polling'))
}

function previewSrc(source: PlaygroundImageSource): string {
  if (source.b64_json) {
    return source.b64_json.startsWith('data:') ? source.b64_json : `data:image/png;base64,${source.b64_json}`
  }
  return source.url || ''
}

async function downloadOutput(source: PlaygroundImageSource, index: number) {
  try {
    await downloadPlaygroundImage({
      source,
      filename: `playground-${imageTaskId.value || 'image'}-${index + 1}`,
      fallback: () => fetchPlaygroundImageContent(imageTaskId.value, index),
    })
  } catch (error) {
    imageError.value = t('playground.downloadFailed')
    console.error(error)
  }
}

watch(userId, (id) => {
  conversations.value = id ? loadPlaygroundHistory(id) : []
  if (!conversations.value.length) startNewChat()
}, { immediate: true })

watch(session, () => {
  syncSelectedGroup()
  syncSelectedModels()
})

watch([selectedGroupId, tab], () => {
  syncSelectedModels()
})

onBeforeUnmount(() => {
  chatAbort?.abort()
  if (pollTimer !== null) {
    window.clearTimeout(pollTimer)
  }
})

onMounted(async () => {
  await loadSession()
})
</script>
