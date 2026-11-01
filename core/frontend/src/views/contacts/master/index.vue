<template>
  <div class="central-contacts">
    <n-alert type="info" :show-icon="true" class="mb-16px">
      Directorio central de correos. Un contacto puede existir sin grupo. La verificación del correo
      y la autorización para campañas son independientes de su registro en este directorio.
    </n-alert>
    <bt-table-layout>
      <template #toolsLeft>
        <span>Mostrar:</span>
        <n-select
          v-model:value="tableParams.filter"
          :options="filterOptions"
          class="w-220px"
          @update:value="resetTable" />
        <n-tag type="info" size="small" :bordered="false">{{ tableTotal }} direcciones</n-tag>
      </template>
      <template #toolsRight>
        <bt-search
          v-model:value="tableParams.keyword"
          :width="290"
          placeholder="Buscar correo, empresa, nombre o cargo"
          @search="resetTable" />
        <n-button @click="fetchTable()">Actualizar</n-button>
      </template>
      <template #table>
        <n-data-table
          v-bind="tableProps"
          :columns="columns"
          :scroll-x="1550"
          :bordered="false">
          <template #empty>
            <n-empty description="El directorio está vacío. Las importaciones futuras se mostrarán aquí, sin requerir un grupo." />
          </template>
        </n-data-table>
      </template>
      <template #pageRight>
        <bt-table-page v-bind="pageProps" @refresh="fetchTable" />
      </template>
    </bt-table-layout>

    <n-drawer v-model:show="showDetails" :width="620" placement="right">
      <n-drawer-content title="Ficha completa del contacto" closable>
        <template v-if="selected">
          <n-descriptions bordered label-placement="left" :column="1" size="small">
            <n-descriptions-item label="Correo">{{ selected.email }}</n-descriptions-item>
            <n-descriptions-item label="Nombre">{{ selected.full_name || '—' }}</n-descriptions-item>
            <n-descriptions-item label="Empresa">{{ selected.company_name || '—' }}</n-descriptions-item>
            <n-descriptions-item label="Cargo">{{ selected.job_title || '—' }}</n-descriptions-item>
            <n-descriptions-item label="Departamento">{{ selected.department || '—' }}</n-descriptions-item>
            <n-descriptions-item label="Industria">{{ selected.industry || '—' }}</n-descriptions-item>
            <n-descriptions-item label="Teléfono">{{ selected.phone || '—' }}</n-descriptions-item>
            <n-descriptions-item label="Ubicación">
              {{ [selected.city, selected.state_region, selected.country].filter(Boolean).join(', ') || '—' }}
            </n-descriptions-item>
            <n-descriptions-item label="Sitio web">
              <a v-if="safeLink(selected.website)" :href="safeLink(selected.website)" target="_blank" rel="noopener noreferrer">{{ selected.website }}</a>
              <span v-else>{{ selected.website || '—' }}</span>
            </n-descriptions-item>
            <n-descriptions-item label="LinkedIn">
              <a v-if="safeLink(selected.linkedin_url)" :href="safeLink(selected.linkedin_url)" target="_blank" rel="noopener noreferrer">Abrir perfil</a>
              <span v-else>{{ selected.linkedin_url || '—' }}</span>
            </n-descriptions-item>
            <n-descriptions-item label="Origen">{{ originLabel(selected.source) }}</n-descriptions-item>
            <n-descriptions-item label="Verificación">{{ verificationLabel(selected.verification_status) }}</n-descriptions-item>
            <n-descriptions-item label="Consentimiento">{{ consentLabel(selected.consent_status) }}</n-descriptions-item>
            <n-descriptions-item label="Suscripción">{{ subscriptionLabel(selected.native_status) }}</n-descriptions-item>
            <n-descriptions-item label="Grupos">
              <n-space size="small">
                <n-tag v-for="group in selected.groups || []" :key="group.id" size="small">{{ group.name }}</n-tag>
                <span v-if="!selected.groups?.length">Sin grupo</span>
              </n-space>
            </n-descriptions-item>
            <n-descriptions-item label="Etiquetas">
              <n-space size="small">
                <n-tag v-for="tag in selected.tags || []" :key="tag.id" size="small">{{ tag.name }}</n-tag>
                <span v-if="!selected.tags?.length">Sin etiquetas</span>
              </n-space>
            </n-descriptions-item>
          </n-descriptions>
          <n-h3 class="mt-20px">Datos adicionales disponibles</n-h3>
          <n-empty v-if="!Object.keys(selected.metadata || {}).length" description="Sin atributos adicionales" />
          <n-descriptions v-else bordered label-placement="left" :column="1" size="small">
            <n-descriptions-item v-for="[key, value] in Object.entries(selected.metadata || {})" :key="key" :label="key">
              <pre class="data-value">{{ printable(value) }}</pre>
            </n-descriptions-item>
          </n-descriptions>
          <n-h3 class="mt-20px">Fuentes y evidencias</n-h3>
          <n-empty v-if="!selected.source_details?.length" description="Sin fuentes adicionales" />
          <n-card v-for="(item, index) in selected.source_details || []" :key="index" class="mb-8px" size="small">
            <pre class="data-value">{{ printable(item) }}</pre>
          </n-card>
        </template>
      </n-drawer-content>
    </n-drawer>
  </div>
</template>

<script lang="tsx" setup>
import { NButton, NFlex, NTag, type DataTableColumns } from 'naive-ui'
import { useDataTable } from '@/hooks/useDataTable'
import { getCentralContacts, type CentralContact, type CentralContactParams } from '@/api/modules/contacts/master'

const { tableProps, tableParams, tableTotal, pageProps, fetchTable, resetTable } =
  useDataTable<CentralContact, CentralContactParams>({
    params: { page: 1, page_size: 20, keyword: '', filter: 'all' },
    fetchFn: getCentralContacts,
    rowKey: row => row.email,
  })

const filterOptions = [
  { label: 'Todas las direcciones', value: 'all' },
  { label: 'Sin grupos', value: 'ungrouped' },
  { label: 'Con grupos', value: 'grouped' },
  { label: 'Directorio maestro', value: 'master' },
  { label: 'Contactos nativos', value: 'native' },
  { label: 'No verificados', value: 'unverified' },
  { label: 'Suscritos', value: 'subscribed' },
]

const showDetails = ref(false)
const selected = ref<CentralContact | null>(null)

const view = (row: CentralContact) => {
  selected.value = row
  showDetails.value = true
}

const originLabel = (source: CentralContact['source']) =>
  ({ master: 'Base central', native: 'BillionMail', both: 'Base central y BillionMail' })[source] || source

const subscriptionLabel = (value: string) =>
  ({ subscribed: 'Suscrito en BillionMail', unconfirmed: 'Sin confirmar', not_subscribed: 'Sin suscripción' })[value] || value

const verificationLabel = (value: string) =>
  ({ verified: 'Verificado', unverified: 'No verificado' })[value] || value

const consentLabel = (value: string) =>
  ({ documented: 'Documentado', not_documented: 'No documentado', opted_out: 'Baja solicitada' })[value] || value

const safeLink = (value: string) => {
  if (!value) return ''
  try {
    const url = new URL(value)
    return ['https:', 'http:'].includes(url.protocol) ? url.href : ''
  } catch {
    return ''
  }
}

const printable = (value: unknown): string =>
  typeof value === 'string' ? value : JSON.stringify(value, null, 2) ?? ''

const columns = computed<DataTableColumns<CentralContact>>(() => [
  {
    title: 'Correo electrónico', key: 'email', minWidth: 230, ellipsis: { tooltip: true },
    render: row => <NButton text type="primary" onClick={() => view(row)}>{row.email}</NButton>,
  },
  { title: 'Nombre', key: 'full_name', minWidth: 150, ellipsis: { tooltip: true }, render: row => row.full_name || '—' },
  { title: 'Empresa', key: 'company_name', minWidth: 190, ellipsis: { tooltip: true }, render: row => row.company_name || '—' },
  { title: 'Cargo', key: 'job_title', minWidth: 170, ellipsis: { tooltip: true }, render: row => row.job_title || '—' },
  { title: 'Teléfono', key: 'phone', minWidth: 150, render: row => row.phone || '—' },
  {
    title: 'Grupos', key: 'groups', minWidth: 170,
    render: row => row.groups?.length ? <NFlex>{row.groups.map(g => <NTag key={g.id} size="small" bordered={false}>{g.name}</NTag>)}</NFlex> : 'Sin grupo',
  },
  {
    title: 'Etiquetas', key: 'tags', minWidth: 150,
    render: row => row.tags?.length ? <NFlex>{row.tags.map(t => <NTag key={t.id} size="small" bordered={false}>{t.name}</NTag>)}</NFlex> : '—',
  },
  {
    title: 'Suscripción', key: 'native_status', minWidth: 130,
    render: row => <NTag bordered={false} size="small" type={row.native_status === 'subscribed' ? 'success' : 'default'}>{subscriptionLabel(row.native_status)}</NTag>,
  },
  { title: 'Verificación', key: 'verification_status', minWidth: 120, render: row => verificationLabel(row.verification_status) },
  {
    title: 'Detalles', key: 'details', width: 100,
    render: row => <NButton size="small" onClick={() => view(row)}>Ver ficha</NButton>,
  },
])
</script>

<style lang="scss" scoped>
.central-contacts { width: 100%; }
.data-value { white-space: pre-wrap; word-break: break-word; margin: 0; font: inherit; }
</style>
