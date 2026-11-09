<script lang="ts" setup>
import { computed, ref, watch } from 'vue'
import { NButton, NInput, useMessage } from 'naive-ui'
import { useUiStore } from '../../stores/ui'
import { t, type MessageKey } from '../../i18n'
import { copyText } from '../../utils/clipboard'
import { decodeText, type DecodeKind } from '../../utils/decode'

const ui = useUiStore()
const message = useMessage()

const input = ref('')
const output = ref('')
const lastError = ref('')

const decodeButtons: { key: DecodeKind; labelKey: MessageKey }[] = [
  { key: 'url', labelKey: 'decode.url' },
  { key: 'base64', labelKey: 'decode.base64' },
  { key: 'hex', labelKey: 'decode.hex' },
  { key: 'html', labelKey: 'decode.html' },
]

watch(
  () => ui.decoderInput,
  (text) => {
    if (text) {
      input.value = text
      output.value = ''
      lastError.value = ''
    }
  },
  { immediate: true },
)

function runDecode(kind: DecodeKind) {
  const res = decodeText(kind, input.value)
  if (!res.ok) {
    lastError.value = t('decode.failed')
    output.value = ''
    message.warning(t('decode.failed'))
    return
  }
  lastError.value = ''
  output.value = res.text
}

function useOutputAsInput() {
  if (!output.value) return
  input.value = output.value
  output.value = ''
}

async function copyOutput() {
  if (!output.value) return
  try {
    await copyText(output.value)
    message.success(t('decode.copied'))
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    message.error(t('ctx.copyFailed', { msg }))
  }
}

function clearAll() {
  input.value = ''
  output.value = ''
  lastError.value = ''
  ui.decoderInput = ''
}

const hasOutput = computed(() => !!output.value)
</script>

<template>
  <div class="decoder-panel">
    <label class="field-label">{{ t('decode.input') }}</label>
    <n-input
      v-model:value="input"
      type="textarea"
      :autosize="{ minRows: 4, maxRows: 12 }"
      class="field-input"
      :placeholder="t('decode.inputHint')"
    />

    <div class="decode-actions">
      <n-button
        v-for="item in decodeButtons"
        :key="item.key"
        size="tiny"
        secondary
        @click="runDecode(item.key)"
      >
        {{ t(item.labelKey) }}
      </n-button>
    </div>

    <label class="field-label">{{ t('decode.output') }}</label>
    <n-input
      :value="output"
      type="textarea"
      readonly
      :autosize="{ minRows: 4, maxRows: 12 }"
      class="field-input"
      :placeholder="lastError || t('decode.outputHint')"
    />

    <div class="footer-actions">
      <n-button size="tiny" quaternary :disabled="!hasOutput" @click="useOutputAsInput">
        {{ t('decode.chain') }}
      </n-button>
      <n-button size="tiny" quaternary :disabled="!hasOutput" @click="copyOutput">
        {{ t('decode.copy') }}
      </n-button>
      <div class="footer-spacer"></div>
      <n-button size="tiny" quaternary @click="clearAll">{{ t('decode.clear') }}</n-button>
    </div>
  </div>
</template>

<style scoped>
.decoder-panel {
  height: 100%;
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: 6px;
  overflow-y: auto;
  padding-right: 2px;
}

.field-label {
  flex: none;
  font-size: 11px;
  font-weight: 600;
  color: var(--text-muted);
  letter-spacing: 0.3px;
}

.field-input {
  flex: none;
}

.decode-actions {
  flex: none;
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.footer-actions {
  flex: none;
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 4px;
}

.footer-spacer {
  flex: 1;
}
</style>
